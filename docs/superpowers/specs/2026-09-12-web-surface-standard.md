# dubplate web surface standard

Status: revision 2.1, 2026-09-12, awaiting Geoff's read. Companion to the
design-language standard
(`docs/superpowers/specs/2026-09-07-design-language-standard.md`, DL
below). DL fixes the theme-variable contract, the two-axis composition,
the three theming tiers, embedded fonts, and the served CSP, and its
section 7 leaves every surface's composition, typography, spacing,
components, forms, feedback states, copy, accessibility practice,
responsive behavior, and code idioms undecided. This standard decides
those, at the level of structure and mechanism.

**Revision 2.** Seven adversarial lenses read revision 1; the fold record
is `docs/superpowers/research/2026-09-12-web-surface-standard-review.md`.
The harvest evidence, the per-clause citations, the reversals list, the
declined table, and the longer arguments behind revision 2's rulings sit
in `docs/superpowers/research/2026-09-12-cairn-harvest-reference.md` (the
reference). Sections 17 and 18 left with them and those numbers are
retired. The five largest folds: the control-edge fix could not win the
cascade as specified and now ships unlayered and quarantined; the
rendered audit engine cannot return a verdict against a client-rendered
bundle, so a port is the primary path and the rendered leg leaves the
per-task gate; the login page's roster-absent answer is fork 4's kind
answer; `themes: false` and the `data-theme` wrapper are gone, because
both contradicted DL; and the no-flash theme pick is a cookie the Go
server reads, not an inline script.

**Revision 2.1.** A prose and fact review of revision 2 returned 29
findings; the fold record is the "Revision 2.1 fold" section of the same
review document. The substantive changes: section 15 now binds every
clause that is a property of a page as built to rung 5c and moves only
new artifacts to rung 7; W5.10's cookie is recorded as a deliberate
supersession of DL R4.2; W13.2's version stamp is a re-measurement
obligation rather than an equality gate across stamps; W2.7's measured
inks become their own layer and only `--color-warning` is claimed as
measured; the display-p3 contrast leg drops to a rung 7 evaluation; and
several uncited specifics (a browser version triple, a SvelteKit 3 API, a
duration) are attributed or removed.

**Owner rulings, Geoff, 2026-09-12.**

1. dubplate borrows cairn's internal design structure: how it layers
   tokens over DaisyUI, contracts its Svelte components, composes pages,
   verifies them, and gates them. It borrows nothing of cairn's look or
   voice.
2. dubplate has its own personality, with room to be a bit more fun than
   cairn.
3. dubplate's design reflects its name, at least gently.

Clauses are numbered `W<section>.<n>` so they never collide with DL's `R`
and `G` series. A clause carrying **Ruled, conductor** is a reading
applied to unblock authoring; section 16 lists the ones that most need
Geoff's eye.

**Every clause outside section 12 carries an enforcement mark, from the
clause itself or from its section's stated default.** `G` gated, naming
the rule. `R` read by a named reviewer at a pass gate. `g` guidance,
binding on an author and checked by nobody. A clause with neither its own
mark nor a section default is a defect the next revision closes. Source pointers are short
names resolving through the reference, which carries the full citation
table and the re-verification procedure. DL, R5, and S20 are dubplate's
design-language standard, rung 5 spec, and 2.0 spec.

## 1. Personality

**W1.1 (structure is borrowed, the look is not).** `g` Every requirement
in sections 2 through 13 is a mechanism. None names a palette value, a
face, or a shipped product string. A measured floor, a threshold, or a
labelled exemplar appears only where the requirement is the number or
where the example is marked as one. Rung 7 authors the look. Src: owner
ruling 1.

**W1.2 (the register, one line).** `R` copy read. dubplate's copy is
plain, warm, and a little playful, written to the Microsoft editor
standard, the same standard rung 5's string deck names. The personality
lives in word choice, headings, empty-state and success-moment copy, the
wordmark and its motif, and the one theme-naming axis W1.3 leaves open.
It never lives in an error message's clarity, a control's accessible
name, or a layout mechanic. Src: owner ruling 2; ADS:56-57.

**W1.3 (the name, gently).** `g` A dubplate is the one-off acetate disc
cut for a sound system to test a track before it is pressed. The product
may wear that lightly, in three homes and no others. Each carries one
illustration rung 7 may discard.

1. *The motif.* The cut disc and its hand-written label, in the wordmark
   and wherever a mark appears. Illustration: a mark built from the label
   rather than the disc, because a circle reads as a generic media glyph
   at 24px.
2. *Vocabulary, where section 8 leaves room.* W8.3's concept nouns and
   W8.5's accessible names bind control names and failure strings, and
   W1.4 refuses personality in failure strings. Headings, empty states,
   success moments, and section labels are open. Illustration: an empty
   library reads "Nothing cut yet"; a finished import reads "Pressed and
   filed".
3. *Theme names, on one axis.* DL R3.6 fixes the spelling
   `<match>-<palette>` and the palette half is a published standard's
   name. The `dubplate` match block's own name, and any palette dubplate
   authors rather than adopts, are the open half.

The word is a motif and never a costume. A first-time member who has
never heard it still reads every page without decoding anything. **Where a clause
in sections 2 through 13 would prevent a personality choice, the choice
wins and the clause is amended**, recorded in section 19 rather than
worked around. The mechanics were harvested from a product with a
different register, and ruling 2 outranks them.

**W1.4 (where personality is refused).** `R` copy read. A reject reason,
a rate-limit line, a security notice, a destructive confirm, and any
string the `internal/reject` contract owns stay plain; the audience is a
person whose upload just failed. The reject strings come verbatim from
the ratified copy deck (`research/passb-copy-final.md`). CLAUDE.md's
"reject wording is contract" rules the verbatim sharing across the
uploader page, the emails, and `upload-check`, not the tone.

## 2. The grammar layer

DL's contract is the palette layer: 21 palette names and 11 structural
names, all DaisyUI's or Tailwind's own. dubplate adds grammar tokens
above it, the layer DL leaves empty. A grammar token names a
relationship, never a palette choice, so it holds across every theme.

**W2.1 (type roles).** `R` svelte-reviewer. A type role is a font size
and its line height, nothing else, declared once outside every theme
block with a paired leading token. Weight, case, tracking, family, and
color stay with the component recipe. The set is small and closed; its
count and values are rung 7's, and each role carries a comment naming
where it is used. Src: ADS:296-323.

**W2.2 (gap roles).** `R` svelte-reviewer. Spacing between related
elements takes the name of the relationship it expresses rather than a
number. Four are fixed here
because every form and list has them: `gap-label` (a label and its
control), `gap-control` (a control and its neighbor), `gap-group` (a
field and the next), `gap-section` (a section and the next). Values are
rung 7's. A two-column field grid uses two relationships per axis, rows
at `gap-group` and columns at `gap-section`, so a reader never mistakes
the right column's first label as continuing the left column's last row.
dubplate ships axis-scoped variants from the start. Src: FORM:45-73.

**W2.3 (container roles).** `G` role-definition test. Properties that
always travel together on one surface are one role. A card shell (radius,
hairline, fill) and a card shadow are two roles, so a surface can take
one without the other. Padding and overflow are never part of a container
role. A role's declarations are copied from the compiled sheet and pinned
by a no-drift test, never hand-approximated. Src:
`admin-css.input.css:186-210`.

**W2.4 (markup writes the role, CSS writes the variable).** `G`
role-utility. A template reaches a grammar token only through its named
utility, never a pixel value, a bracketed `var()` wrapper, or a bare
Tailwind size step for a role the grammar covers. A scoped `<style>`
references the custom property directly, since CSS is not markup. Each
`type-*` utility sets font-size and line-height and reads Tailwind's
leading variable first, so an explicit `leading-*` still wins. The
role-utility rule replaces cairn's `gap-scale`, whose subject this clause
forbids outright. Src: ADS:305-323.

**W2.5 (every token reaches markup as a utility).** `G` role-definition
test. A grammar token is declared in Tailwind 4's `@theme` block **and**
exposed through an `@utility` definition, unless its name falls in a
Tailwind namespace that already emits one. `--gap-*` is such a namespace;
`--type-*` is not, and a `--type-body` key emits no CSS and no build
error. The definition test asserts every role's utility is in the built
sheet. Src: `admin-css.input.css:170-205`.

**W2.6 (the palette and grammar boundary).** `G` DL G8. DL's operator
branding tier re-tunes palette names freely and never declares a grammar
token; DL G8 already gates that stylesheet's name set, which forecloses a
grammar token there without a new rule. DL's operator CSS file is the
adopter's, outside the contract and outside every gate's reach, and
W5.11's literal fallback is the only defense there. Src: DL R4.3, R4.4,
G8.

**W2.7 (a fill is not an ink).** `G` contrast-extension. DaisyUI tunes
its status colors as fills. cairn measured `--color-warning` as small
text on the base surface at about 2.2:1 (ADS:207-215), which is the
evidence for the split; `--color-success` and `--color-error` are the
same mechanism and are measured at rung 7 before their inks are set. Each
status color therefore has two tokens, the DaisyUI fill and a separately
measured on-surface ink, and dubplate's own markup reaches the ink
through a named utility, never `text-warning`. The ink defaults to the
fill.

This is additive to DL R2.6 rather than a replacement. R2.6's third
condition covers the paths an ink utility cannot intercept, DaisyUI's own
`-soft` modifiers, `text-*` utilities, and `.validator-hint`, and a
palette that cannot clear the floor still records the exception G9 parses
in DL's fixed form. R2.6's third condition and its exception line are
unchanged by this standard; revision 1 read R2.6 as a replacement and
that reading is withdrawn.

The ink tokens are neither palette-block declarations nor grammar tokens.
They are a third layer, the measured status inks, living in a per-palette
ink block in dubplate's own stylesheet and selected by the composed
`data-theme` value. DL R2.1's 21-name palette block and section 2's
palette-free grammar layer both stand: a measured ink is a palette
choice, so it cannot be a grammar token, and DL G2 fails on a
twenty-second name in the palette block.

The 4.5:1 figure is SC 1.4.3's normal-text floor. 1.4.3 allows 3:1 at
large text and dubplate holds every status ink to 4.5:1 regardless,
because a status ink's type role can change under it. Src: ADS:207-221.

**W2.8 (the inventory has one home).** `G` role-definition test. The
grammar token list is one module that the definition test, the
styleguide, and the static gate all read, so the inventory cannot drift
across call sites. cairn's third reader is a reference docs page;
substituting the styleguide is dubplate's own. Src:
`grammar-tokens.ts:1-8`.

**W2.9 (fluid type, one mechanism per range).** `g` Where a type step
uses `clamp()`, exactly one fluid mechanism is active per width range. If
the root font size is itself a clamp (W4.9), each step's ceiling is
rescaled so it does not compound and each floor keeps its phone value. A
`clamp()` whose minimum exceeds its maximum resolves to the minimum
unconditionally; a step that never varies is written as a constant. Src:
PDS:108-135.

**W2.10 (three Tailwind 4 traps).** `G` spacing-collision, plus review
for the other two. Each was found in production and each costs one line
to avoid. All three are version-coupled under W13.2.

- A `--spacing-<k>` key shadows `max-w-<k>` for every `k` Tailwind's
  `--container-*` set also defines, because both namespaces feed one
  utility prefix and spacing is read first. The default container set
  runs `3xs` through `7xl`, so `--spacing-md` is as exposed as
  `--spacing-2xl`. A static rule asserts no `@theme` `--spacing-*` key
  collides with the pinned Tailwind's container set. Src: WM
  `chassis/README.md:84-95`.
- `text-2xl` sets font-size and line-height; `text-[1.5rem]` sets
  font-size alone, so converting casually drops a leading. Write
  `text-[1.5rem]/[2rem]`, and prefer a type role. Src: ADS:340-344.
- An unlayered rule beats every Tailwind utility, because cascade layers
  resolve before specificity. `margin: 0 auto` silently zeroes that
  element's own `mt-*`; write `margin-inline: auto`. Src: WM
  `chassis/README.md:96-110`.

## 3. Page composition

**W3.1 (name the contexts).** `g` Each surface belongs to a named context
with its own chrome budget. cairn has two, office (lists and settings,
persistent chrome) and desk (an open document, chrome recedes). dubplate
names its own at rung 7; the uploader and the later player are the desk
candidates. What transfers is that a route's chrome is decided by its
context, never per page. Src: ADS:122-128.

**W3.2 (content width).** `R` visual-verifier. Every non-desk surface
caps its content width and centers it, so a wide screen never stretches a
list row into a dead band between its title and its date. A list may
compose narrower than the shell cap when its content is short. Header,
footer, and main share one measure. Src: ADS:155-162, ADS:471,
PDS:157-162.

**W3.3 (page header, on the screens that have one).** `G` h1-count, plus
`R` svelte-reviewer. A list or settings screen mounts the same page
header: an optional eyebrow, one `h1`, an optional muted meta line, and
an optional single action, which sits beside the `h1` and never trails
the content. Search belongs to a toolbar below the header band. A count
line has two homes, the toolbar's count line or the header's meta, at one
type role, never both.

A single-purpose form screen (rung 5's login and consumption pages) and a
direct-manipulation screen (the uploader's drop surface) each declare
their own anatomy: the primary action is inside the form or is the
surface itself, and neither mounts the header's action slot. The part
binding every page is the `h1` count, which is what the gate asserts;
W7.1's heading-level prop is how a screen with no header still has one.
Src: TK `PageHeader.svelte:3-14`, ADS:459-462.

**W3.4 (the topbar pins its height).** `R` visual-verifier. The topbar
pins its height explicitly, because a content-driven navbar drifts with
font metrics and two border-bottoms stop meeting at the seam. Sidebar and
topbar share that height. Chrome depth, fill, and hairline are the
theme's: DL R2.1 puts `--depth` and `--noise` on the structural axis, so
flatness is a theme author's choice and not this standard's. Src:
ADS:1024-1031; DL R2.1.

**W3.5 (navigation is flat).** `R` svelte-reviewer. A category header
costs the reader a decision on every visit, a cost a small navigation
never earns back. The way back never disappears: below the persist
breakpoint the chrome becomes an overlay drawer, and a toggle and its
target share one breakpoint literal so the toggle never dangles beside a
persistent sidebar. Src: ADS:363-366, ADS:139-151.

**W3.6 (skip link and focusable main).** `G` skip-link assertion. The
skip link is the first focusable element and targets
`<main id="main" tabindex="-1">`, which is SC 2.4.1. The `tabindex` is
load-bearing everywhere: no browser moves focus to a non-focusable
fragment target, and WebKit additionally does not set the sequential
focus navigation starting point (bug 17450). `main:focus` suppresses the
outline there alone. A sticky header sets `scroll-padding-top` from the
header's own tokens, for SC 2.4.11. Src: PDS:153-157, WM `site.css:46-53`.

**W3.7 (404s, and who answers them).** `R` web-auth-security-reviewer.
Two layers answer a 404 and the Go mux answers first. A path under the
owned prefix
matching no route returns 404 from the Go mux and never reaches the
fallthrough. That terminator is rung 5's and it is a trust boundary, and
nothing here weakens it. Outside the prefix the streamer owns everything.

The SvelteKit app's own 404 is a client-router state, reached by
client-side navigation to a path the app owns and does not have. The root
`+error.svelte` renders it, importing its own stylesheets and chrome,
because for an unmatched path a route group's `+error.svelte` never
mounts; a thrown in-route `error()` still reaches the group boundary,
which is a different code path. The 404 copy is one human sentence; any
other status shows the error's message or a plain fallback.

DL R4.5 states the serve-layer mechanism in the singular, one documented
rewrite over the embedded `index.html`, and that single rewrite is what
the Go layer runs. Where prerendering emits a document per owned route,
the serve layer applies that same rewrite to whichever document it
returns and DL R5.6's hash set covers every inline script in each. That
widens R4.5's subject from one file to a set, which section 19 records.
Src: WM `chassis/README.md:169-199`; R5:905-907.

**W3.8 (the main column's width bug).** `R` svelte-reviewer. A
`flex: 1 1 0%` main whose descendant is intrinsically wider (a nowrap
table, an `overflow-x: auto` box) shrink-to-fits instead of stretching,
and `min-width: 0` alone does not fix it: the main column carries both
`width: 100%` and `min-width: 0`. dubplate's verdict tables are exactly
that descendant. Src: WM `composition.css:117-128`.

**W3.9 (composition primitives, few and real).** `R` svelte-reviewer. A
composition recipe declares its own `--<primitive>-*` custom properties
as the first lines of its rule, then consumes them, so one instance
overrides one property without touching the shared scale. Section rhythm
is one `--flow-space` variable block elements locally redeclare, never a
margin per element. A heading takes large space above and tight space
below, so it belongs to the text it introduces. dubplate ships fewer
primitives than cairn, each with a real call site in its comment. Src: WM
`composition.css`, `prose.css:71-112`.

**W3.10 (a breakpoint is not a token).** `g` A media query cannot read a
custom property, so a breakpoint is a fixed value and its comment says
so. Src: WM `composition.css:99-103`.

## 4. Responsive behavior

**W4.1 (the five-viewport bar).** `G` viewport-overflow, plus `R`
visual-verifier. Every surface is composed at 320, 390, 768, 1440, and
2560 CSS pixels. Composed means composed at every width, not merely
unbroken. The mechanical floor sweeps 390 and 320; the composition read
at all five is a grader's or a person's. The both-schemes half binds from
rung 7, because R5's R36 gives rung 5 one composed theme. Src: ADS:28-29.

**W4.2 (responsive by construction).** `g` Responsive behavior lives in
the shared components, not in each screen, so a screen built from them
recomposes at 320 and 2560 without its author having thought about
either. A ruling that exists only in this document has not finished
shipping; it lands as a component, a token, or a gate. That is the
standard this document holds itself to, and the enforcement marks are how
a reader checks it. Src: ADS:43-44, ADS:45-49.

**W4.3 (a phone is a full device).** `g` A member checks an upload
verdict from a couch. The uploader and the verdict list are designed for
the phone context, not shrunk to it. Src: ADS:58-60.

**W4.4 (a row declares its own breakpoint).** `R` visual-verifier. A
multi-element row never relies on flex-shrink alone, because a flex
item's default `min-width: auto` refuses to shrink below its content
width, so the row collides or wraps mid-phrase instead of reflowing. The
row states a breakpoint below which it stacks into a column, confirmed by
rendering at the narrow width, never by reading the classes. A class that
merely lets text wrap is not a layout that reflows. The stacked gap is
rung 7's choice; cairn's two passages disagree on it. Src: CRAFT:87-97,
CRAFT:137-144.

**W4.5 (tables fold columns, and a scrollbar is the last resort).** `R`
daisyui-a11y-reviewer. A data table keeps the one column a reader
identifies the row by, hides every other column below the breakpoint, and
folds each hidden value into the identifying cell as one meta line, so no
information disappears, it moves. The recipe ships as literal markup,
because cairn's acceptance trials failed twice on the prose rule and
passed only on a copyable skeleton. The column names are placeholders,
not shipped strings.

```svelte
<table class="table">
  <thead>
    <tr>
      <th scope="col">Album</th>
      <th scope="col" class="hidden sm:table-cell">Artist</th>
      <th scope="col" class="hidden sm:table-cell">Tracks</th>
      <th scope="col"><span class="sr-only">Actions</span></th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <div class="type-subtitle font-medium">{album.title}</div>
        <div class="type-meta sm:hidden">{album.artist} · {album.tracks} tracks</div>
      </td>
      <td class="hidden sm:table-cell">{album.artist}</td>
      <td class="hidden sm:table-cell">{album.tracks}</td>
      <td>{@render actions(album)}</td>
    </tr>
  </tbody>
</table>
```

A horizontal-scroll wrapper is containment, necessary but not sufficient:
any flex or grid ancestor between the wrapper and the surface edge
carries `min-width: auto` and lets the content push the page wider than
the viewport. A scrollbar earns its place only after every foldable
column is folded and the table still does not fit.

The table stays `display: table`. Engines once dropped table semantics
under `display: block`, which was true until Chrome 80, Firefox 62, and
Safari 17 and is now false, so that is not the reason. The reason is that
`display: contents` still drops them across engines and that the folding
recipe needs the table's own layout. The scroll wrapper is a separate
element carrying `role="region"`, an `aria-label`, and `tabindex="0"`
applied only while the content overflows, so a keyboard user can reach it
without meeting a dead tab stop on a wide screen. It is also SC 1.4.10's
data-table exception. Src: CRAFT:98-109, CRAFT:145-224; Roselli.

**W4.6 (ledger rows stack).** `R` visual-verifier. A multi-column row
that is not a table recomposes below its breakpoint into two lines: the
primary control full width, then a self-labeling muted meta line. The
header band then keeps only the primary column's label. Src: ADS:833-840.

**W4.7 (dialogs on a narrow screen).** `R` visual-verifier. See W5.14 for
the dialog's own mechanics. A modal never fills the viewport on a normal
screen: the box sizes to its content under a cap on both axes, so the
backdrop frames it, and the box is the scroll container, with header and
footer holding while the body scrolls. The cap's values are rung 7's.
Below the narrow breakpoint a dialog may become a full-height or bottom
sheet, and only there. A two-pane dialog collapses to one column on a
narrow screen. Src: ADS:494-501, ADS:518.

**W4.8 (the header lockup without JavaScript).** `G` contents-wrapper
assertion. Below the header's breakpoint the nav-and-toggle wrapper takes
`class="contents"`, leaving the box model so nav and toggle rejoin the
outer flex row as flat siblings that `order-*` arranges into a deliberate
two-row lockup; at the breakpoint it becomes a flex box again and
`md:order-none` restores source order. No hamburger, no duplicate mobile
markup. The `contents` element is a bare `<div>` with no role, landmark,
or accessible name, because implementations still remove such an element
from the accessibility tree and some remove its descendants; it is never
applied to a list, a table part, a landmark, or the `<nav>` itself. The
rendered gate asserts the nav's links are in the accessibility tree below
the breakpoint. Src: WM `SiteHeader.svelte:26-27`, `:106-111`.

**W4.9 (the ultrawide posture).** `R` visual-verifier. Above about 1440px
the root font size scales through one `clamp()` on `html` and holds at
its ceiling. Because every dimension is rem-based, the measure included,
one value grows the whole surface proportionally rather than reflowing
into new columns. W2.9 keeps this from compounding with the type steps.
This also discharges SC 1.4.4. Src: PDS:135-146.

**W4.10 (touch targets, two floors).** `G` touch-targets. Two floors
exist, they are different criteria, and they are not the same number.

44px is the requirement's word, SC 2.5.5 at AAA. DL R3.7 carries it on
the structural axis through `--size-field` so a plain `.btn` renders at
44 or more, and DL G2 gates it statically. A control reaches the target
by growing its box invisibly (`min-height: 2.75rem` on an inline-flex
control), never by inflating its padding and visual weight. A
checkbox-in-label row grows its hit area with one scoped `:has()` rule
that adds padding and refunds it as negative margin, so the target clears
with zero flow change. `.btn-sm` keeps its own `min-h-11` at each call
site, as DL R3.7 requires.

24px is the measured floor, SC 2.5.8 at AA. dubplate measures it because
a token check cannot see a rendered 20px checkbox or a label that
activates it. cairn's own reason for lowering its gate from 44 to 24 does
not transfer, since DL R3.7 forbids the condition that produced it.

**The measurement is a square, not an area**: a 24 by 24 CSS pixel
axis-aligned square must fit wholly inside the activation region, the
control's box unioned with every label the platform reports as activating
it and with a qualifying `::before` expansion. **The rule carries SC
2.5.8's five exceptions or it is not the criterion.** Inline (a target in
a sentence, or sized by the line-height of non-target text) and Spacing
(undersized targets whose 24px-diameter circles do not intersect) are
implemented before anything fails; Equivalent, User Agent Control, and
Essential are suppression grounds with a named reason. Without them the
rule fails conforming inline links on the first page with a sentence in
it, which W12.2 rules a gate-integrity defect. The sweep runs at every
W4.1 viewport, because a control can be 24px on a phone and 20px in a
desktop-only toolbar.

**Ruled, conductor 2026-09-12:** the rendered rule (W12.4's
`touch-targets` row, at W12.2's error tier) fails below 24 on the
activation region and reports below 44 as advisory. Geoff decides whether
44 becomes an error tier once rung 7's component set exists. Src: RULES
`rendered/touch-targets.ts`; WCAG 2.2.

**W4.11 (tap acknowledgment).** `g` `-webkit-tap-highlight-color` is
dampened to a low-alpha mix of the base content color, never removed,
because a segmented control segment or a text-link button carries no
styled `:active` state and the tap still needs acknowledgment. Src:
ADS:196-200.

**W4.12 (vertical alignment).** `R` visual-verifier. Two runs of type
share a baseline; anything else shares an ink center, never a box center.
The carve-out is the case this document meets most often: a padded chip
is itself a box, its border and fill draw ink at its own edges, so its
padding-box center is the reading, and the type inside it is a separate
optical question. The optical offset is recorded as a number in the
recipe once the call is made. Src: ADS:1054-1066, PDS:303-315.

## 5. Component contracts

These rules bind every `.svelte` file under `web/src`. The
`svelte-conventions` and `ts-conventions` skills govern the comment
shape; this section governs the contract. Unless a clause says otherwise
its mark is `R`, read by `svelte-reviewer` and `daisyui-a11y-reviewer` at
each pass gate.

**W5.1 (the comment states the contract and the refuted alternative).**
The `@component` block says what the component is, what it owns, what it
deliberately does not own, and what failure its existence prevents. Where
a visual value was chosen over another, the comment names the alternative
and why it was refuted, with the measurement. "We chose X" without "we
refuted Y because Z" is incomplete; the next author re-derives Y. Src: TK
`AdminTable.svelte:1-24`.

**W5.2 (props are data plus snippets).** A component takes plain data and
snippets the caller authors, and carries no domain knowledge. A table
component takes no `rows` prop; the caller's markup stays its own, and
the component carries the nowrap floor and the overflow fallback a
consumer would otherwise rediscover. A snippet that renders a datum is
typed `Snippet<[T]>` and receives it. Src: TK `AdminTable.svelte:10-12`.

**W5.3 (interactive primitives are fully controlled).** A primitive that
can be open, expanded, selected, or paged takes its state and an
`on<Thing>Change` callback as props and owns no `$state` for it. It never
decides whether it may open; it reports when it wants to change.
Coordination a primitive cannot own (one row expanded at a time) lives in
the composing screen. `createEventDispatcher` is not used. Src: TK
`ExpandableRow.svelte:8-14`.

**W5.4 (runes discipline).** `$props()` is destructured once against a
named `interface Props` with a TSDoc line per prop stating the contract,
never the type. `$derived` computes everything renderable; `$effect` is
never used for derivation, and where a reader would wonder why, the
comment says so. `$state` is for genuine local UI state; `$state.raw`
where a value is replaced wholesale; a plain `let` for a DOM reference
read only imperatively. An `$effect` returns its cleanup and guards a
stale async resolution with a cancelled flag. `$props.id()` mints a
per-instance id for `aria-controls`. DOM-lifecycle behavior uses
`{@attach}`, Svelte 5's successor to actions, with `fromAction` for a
library shipping only an action. Src: CI:142-145.

**W5.5 (one element with a conditional class).** When a mode flips live,
the element stays and its class changes. A two-branch `{#if}` destroys
and recreates the element, dropping focus and any in-progress input
composition inside it. Src: TK `FieldLabel.svelte:57-61`.

**W5.6 (pure logic leaves the component).** Arithmetic, formatting, and
any rule a test can state lives in a plain `.ts` module beside the
component, so the unit project runs it in Node without mounting. Export
surface is pruned by evidence: a module no consumer reaches leaves the
barrel. Src: TK `pagination-window.ts:1-6`.

**W5.7 (DaisyUI composition).** A component's comment names exactly which
DaisyUI classes it assembles from, so the blast radius of a renamed class
is a grep. Layout, truncation, min and max width, and wrapper geometry
live in the scoped `<style>`, used only for what a utility cannot express
(a `color-mix` ground, a backdrop filter, a descendant selector). A
scoped block never reaches for a DaisyUI internal selector; those live in
W5.22's file. Src: TK `AdminTable.svelte:8-10`.

**W5.8 (layers, and what an unlayered rule is for).** `G`
unlayered-count, unlayered-font-clobber. The measured layer order is
`properties < theme < base < components < utilities`, with DaisyUI's
component rules in nested sublayers of `utilities` and a layer's own
direct content beating its sublayers. So `@layer components` is the home
for a rule that must lose to a Tailwind utility, and it cannot beat
DaisyUI at all. A rule that must beat DaisyUI is unlayered by necessity. Such a rule is a
forced exception: it is pinned by exact selector, counted by the gate as
a ratchet, and carries the forcing reason in its commit. W5.18's control edge,
W7.8's guarded control, W5.13's `.menu` hidden rule, and W9.3's
reduced-motion kill switch are the standing entries, all in W5.22's file.
A scoped `<style>` never declares `font-family`, `font-size`, or
`font-weight` on an element that also carries a font utility, because the
scoped block is unlayered and wins regardless of specificity. Src:
ADS:76-83, AUD:80.

**W5.9 (the theme element).** `G` DL G7, for the precedence half alone;
`R` svelte-reviewer for the rest, since DL G7 asserts link order,
`no-cache`, and one computed-value precedence check and sees neither the
styled-class rule nor the absent guard block. `data-theme` lives on the
document element, per DL R4.2, stamped by the Go binary into the document
it serves, and the theme element carries no styled classes. It does not
live on an inner wrapper: DL R4.1's tier precedence works only because
DaisyUI's `[data-theme]` block and the branding stylesheet's `:root` are
equal specificity on the same element, and a wrapper one level down
shadows every inherited palette variable unconditionally, killing DL's
tiers 2 and 3. DaisyUI's built-in themes are not disabled, because DL G10
and R3.4 require a test reading a computed structural variable while a
built-in palette is active; how the built-ins reconcile into the two-axis
model stays rung 7's detail, as DL left it. There is no
`:root:not([data-theme])` guard and no `prefers-color-scheme` fallback
block: DL R4.5 stamps the attribute on every served document, so the
guard would be dead code, and at (0,2,0) it would outrank the branding
stylesheet at any link order. A system-following pick is a third value
the Go layer resolves and stamps. Src: DL R4.1, R4.2, R4.5, R4.6.

**W5.10 (the theme pick never flashes, and no inline script ships).** `R`
web-auth-security-reviewer, plus `G` a served-document assertion that no
`<script>` is present beyond SvelteKit's own generated set. DL G6 does
not gate this clause: a hand-added inline theme script carrying a
matching hash passes G6, and neither G6 nor G7 sees the cookie stamp at
all.

The server stamp is the whole mechanism. For a signed-in member DL R4.5
stamps the stored pick; for a signed-out visitor the
toggle writes the pick to a cookie (`SameSite=Lax`, readable by the
client for its own state) and the Go layer stamps `data-theme` from that
cookie on the document it already rewrites. With no cookie the operator's
configured default is stamped. The document is already served `no-cache`,
so a per-visitor stamp costs nothing new, and the pick is lost only with
the cookie jar, which is the same pre-rung-6 posture DL R4.2 accepts.

**This supersedes DL R4.2's ruled `localStorage` store for the signed-out
pick**, and section 19 records it. The reason is the inline script: a
`localStorage` pick can only be read after the bundle runs, so removing
the flash needs an inline script in the served document, and DL R5.6's
CSP hashes only the scripts SvelteKit generates. The cookie is readable
by the Go layer while it already rewrites the document, so the flash goes
without a script the policy would block. `localStorage` is not used for
the pick.

No inline theme script ships. The browser would block a hand-added one
even though the Go header allowed it, because SvelteKit hashes only the
scripts it generates and never parses `app.html`, so the two policies
intersect and the wrong theme flashes on every load. The toggle is a
standing override from the first explicit choice on, never a tri-state
control. A cross-fade is scoped to the click that flips the theme and to
the chrome surfaces alone, never to first paint, a media change, or an
image. Rung 5 builds no picker and authors one composed theme, so this
clause lands at rung 7. Src: DL R4.2, R4.5; R5:1329-1332.

**W5.11 (a literal fallback under every token-derived color).** A
declaration reading a token the theme might not define carries a literal
fallback first, so a misplaced component is visibly something rather than
visibly nothing. This is the only defense reaching DL's operator CSS
tier. Src: TK `StatusChip.svelte:144-151`.

**W5.12 (the disclosure primitive, built once).** Any dropdown, facet, or
menu is one primitive carrying the five dismissal mechanics:
`aria-expanded` and `aria-controls` in sync; focus into the panel on
open; Escape closes and returns focus to the trigger, container-scoped so
a nested disclosure closes only itself; an outside pointerdown closes
without moving focus; focus leaving the boundary closes without moving
focus, guarded by `document.hasFocus()` so a window blur does not read as
a Tab out. For a non-popover panel the `hidden` attribute is the primary
hiding mechanism, never a class; a popover's hiding belongs to the user
agent. The first-focusable query excludes `[tabindex="-1"]`. Src: TK
`ToolbarDisclosure.svelte:179-276`.

**W5.13 (the focus-driven open method is refused, not the class).** A
menu opens through the platform Popover API or through W5.12's primitive.
What is refused is DaisyUI's CSS-focus open method, which opens on
focus-in-transit and ignores Escape; `.dropdown` is a style class and
DaisyUI's own popover example uses it, so the refusal keys on the
mechanism, not the class name, and W12.3's rule does the same. The
invoker's `aria-expanded` comes from the user agent for a `popovertarget`
button and is not authored by hand, where a copy can desync during a
light dismiss; a trigger whose panel is a `role="menu"` carries
`aria-haspopup`, and `aria-controls` is belt and braces. Where W5.12's
primitive wraps a DaisyUI menu, `.menu`'s author `display: flex` outranks
the UA `[hidden]` rule and needs its own hidden rule in W5.22's file. CSS
anchor positioning is not yet Baseline across current evergreen browsers.
Its support is checked at rung 7 against the table current then, and the
clause carries a `WATCH:` comment under W10.7 naming that table rather
than a version triple this document cannot keep current. Src:
ADS:526-530; MDN Popover API.

**W5.14 (dialogs).** A modal is a native `<dialog>` opened with
`showModal()`, which gives the focus trap and Escape for free;
`aria-modal` is not authored on top of it. A standalone dialog component
exports `open()` as its contract rather than hosts reaching through
`bind:this`. Its focus lifecycle is two shared functions (remember the
origin, refocus it), not a stateful factory; the open path focuses Cancel
after `tick()`, and a programmatic `.click()` passes its origin
explicitly because it does not focus its target. A dialog holding its own
`<form>` mounts outside any page form, because a nested `<form>` start
tag is ignored by the parser while a form element pointer is already set,
so the dialog's controls silently join the page form instead. Src: CI:146-150,
ADS:489-493.

**W5.15 (roving tabindex).** A pick-one filter is a real
`role="radiogroup"` over `role="radio"` with `aria-checked`, never
`aria-pressed`. The checked option is the only tab stop; with nothing
checked the first option is. Arrows move selection and focus together.
Home and End at the edges are dubplate's addition, not the APG radio
pattern. A check glyph is the non-color cue. A grid or listbox is one tab
stop with arrows, Home, End, and Space or Enter. The roving index is
clamped in its accessor, not synced by an effect, so a shrinking list
never leaves the group without a tab stop. A sortable column is a
`<th scope="col">` carrying `aria-sort`, whose header text is wrapped in
a `<button>` that toggles it; `aria-sort` is dropped on a `button`, so a
static rule asserts it never appears on a non-header element. Src:
ADS:902-912; ARIA APG.

**W5.16 (lists keep their role).** `G` list-role. Tailwind's Preflight
writes `list-style: none` on every `ol`, `ul`, and `menu`, so every list
is marker-suppressed before dubplate writes a line of CSS, and WebKit
with VoiceOver drops list semantics from a marker-suppressed list.
dubplate keeps Preflight, because the alternative is a hand-maintained
reset with a per-bump re-verification chore. So the rule is affirmative,
not a prohibition: every `<ul>` or `<ol>` conveying a list carries
`role="list"`, and each `<li>` additionally carries `role="listitem"`
only where the item's `display` is changed (`flex`, `grid`, `.list-row`,
`.menu`), which is the separate defect. A list inside `<nav>` keeps its
semantics in Safari without the role, and carrying it anyway costs
nothing. A bulleted list under Preflight needs `list-style-type` and
padding restored, not a padding reset. Src: ADS:595-615; Tailwind
Preflight; O'Hara.

**W5.17 (refused DaisyUI defaults).** `G` refused-defaults. These stock
patterns are replaced, because a builder reaching for a default reaches
for what DaisyUI's docs show. Each row carries the version its claim was
measured against, and each claim is a falsifiable test rendering the
stock default and asserting the failing measurement, so a DaisyUI that
fixes one turns the test red and the row is re-argued rather than
outliving its reason.

| Default | Why refused (measured against) | Use instead |
|---|---|---|
| `badge-ghost`, `badge-outline`, `badge-soft`, `badge-dash` | Hardcoded ground or `currentColor` border melts into a row or zebra stripe; none clears 3:1 on its edge in both schemes (5.7.20) | The chip register set (W7.10) |
| The CSS-focus dropdown open method | Opens on focus-in-transit, ignores Escape, fights `aria-expanded` (5.7.20) | W5.12 and W5.13 |
| Native `disabled` on a guarded control | Removes the reason from the tooltip and tab order; 5.6 added `[aria-disabled="true"]` to the **`.btn`** disabled selector with `pointer-events: none` (5.6) | W7.8; native `disabled` for mid-submit only |
| A flat `base-300` card border | Invisible against the ground in one scheme (5.7.20) | The theme-adaptive hairline and shadow pair |
| `text-warning`, `text-success`, `text-error` as text, and `.validator-hint` | Resolve to the fill tone; `text-warning` measures about 2.2:1 as small text and the other two are measured at rung 7 (5.7.20) | W2.7's ink utilities |
| `.list` on a bulleted list | Forces flex and 14px (5.7.20) | A restored list style (W5.16) |

cairn's ported rule covers four of these by name: the ghost badge, the
bare dropdown, the native `disabled` (keyed on a cairn marker class
dubplate renames), and the flat card border. The outline, soft, and dash
badges, the `text-*` family with `.validator-hint`, and the `.list` row
are what dubplate's port adds. Src: AUD:73, ADS:843-849.

**W5.18 (the DaisyUI control-edge fix).** `G` unlayered-count, plus a
rendered border-contrast assertion. DaisyUI 5's unchecked `.checkbox` and
`.radio` and unfocused `.input`, `.select`, and `.textarea` resolve their
border through a 20% mix measuring 1.49:1 light and 1.77:1 dark, under
SC 1.4.11's 3:1
non-text floor. The fix raises the edge to a 55% mix, measured at 3.59:1
light and 4.96:1 dark against DaisyUI 5.7.20; `.toggle` measured 3.12:1
and 4.37:1 and needs no change. The same mix serves the scrollbar thumb
and the outline chip border. **The rule is unlayered by necessity and
lives in W5.22's file**: the border it overrides is a utilities-sublayer
declaration, so an `@layer components` rule cannot outrank it at any
specificity and the fix would compile, pass every static rule, and change
nothing on screen. It is one of W5.8's pinned exceptions and one of
W13.2's version-coupled values. DL's contrast gate measures text pairs
and cannot see this. Src: ADS:222-238.

**W5.19 (element ids).** `G` element-id. An element id follows
`dubplate-<component>-<element>`. Src: CI:91.

**W5.20 (the smooth-scroll trap).** `scroll-behavior: smooth` for anchor
jumps also animates the router's post-navigation `scrollTo`, turning
every route change into a glide. The root layout brackets the router's
scroll window with a class set in `beforeNavigate` and cleared in
`afterNavigate`, with a `navigation.complete.catch` fallback for a
cancelled navigation. Src: WM `site.css:55-81`.

**W5.21 (a scan-root miss compiles to nothing).** `G` uncompiled-class. A
Tailwind class used only outside the stylesheet's scan roots produces no
CSS and no build error. dubplate keeps Tailwind 4's automatic source
detection, which scans the project by default; cairn's reason for
disabling it does not transfer. Under automatic detection this clause
reduces to the uncompiled-class gate, which is the net either way. Src:
ADS:345-349.

**W5.22 (DaisyUI-internal overrides live in one file).** `G`
unlayered-count. Every rule naming a DaisyUI internal selector, or that
must outrank a DaisyUI declaration, lives in one file
(`web/src/lib/styles/daisyui-overrides.css` or equivalent), each carrying
the DaisyUI version it was measured against and the measurement in its
comment. Component scoped styles never reach a DaisyUI class. Today's
entries: W5.18's control edge, W7.8's guarded-control `pointer-events`,
W5.13's `.menu` hidden rule, and W9.3's reduced-motion kill switch. A
DaisyUI bump then reads one file rather than sweeping `web/src`.

**W5.23 (the coupled control families reach markup through a recipe).**
`G` refused-defaults, extended to flag the raw class outside the recipe
file. Every control family carrying a refused default, an edge fix, a
guarded pattern, or a measured floor is reached through one named recipe:
a component where the recipe has behavior, a documented class string
constant where it does not. No screen writes that family's DaisyUI
classes directly. The families are roughly six to nine (button, input,
select, textarea, checkbox, radio, badge or chip, menu, dialog) and the
count is stated so the set cannot accrete. This is deliberately narrower
than a full wrapper set: cairn retired three such wrappers for zero
consumers.

## 6. Forms

Unless a clause says otherwise its mark is `R`, read by
`daisyui-a11y-reviewer` at each pass gate.

**W6.1 (three label levels).** `R` daisyui-a11y-reviewer. A label's level
is chosen by what it groups, never by copying the nearest field. A group
legend names several controls: the eyebrow recipe on a `<legend>`. A
field label names one control: sentence case, medium weight, base content
ink, no tracking, no uppercase, because a field label at the group
title's weight reads as a second title. An inline control-adjacent label
sits beside a checkbox or a compact control: body role, muted ink, plain
weight. Src: FORM:13-43, ADS:263-284.

**W6.2 (the label wraps the control and nothing else).** `R`
daisyui-a11y-reviewer. A wrapping `<label>` adds every piece of text
inside it to the field's accessible name, so only the label text sits
inside. A hint or an error line sits outside the label and is wired by
`aria-describedby`, because a hint inside would join the accessible name
and then be announced again as the description. Where layout forces a
hint inside, the label text takes its own element and the control carries
`aria-labelledby` pointing at it, with the reason in a comment. A
wrapping label with no `for` associates with only its first labelable
descendant, so every later control in a compact row needs its own label
or `aria-label`. Src: TK `FieldLabel.svelte:17-30`; FORM:114-118.

**W6.3 (stacked by default).** `R` A stacked label is the default
register and the only one used in a multi-column grid. An inline label is
the verified exception: at mid-desktop widths a `<span>` label is the
more shrinkable child of its row and wraps inside a narrow fixed track
once a second column opens, a width no automated rule samples. Src:
FORM:82-129.

**W6.4 (the submission row).** `R` A form with a discardable draft or a
dismissable context ends with a plain Cancel beside a filled primary
submit, `gap-control` apart. It is the form's one filled action. **A
single-action form ships the submit alone**: a form with one field and
one control has nothing to cancel to, which is the shape of rung 5's
login page and its consumption interstitial. The row's alignment is rung
7's visual choice, not a mechanic. Src: FORM:131-140; R5:1393-1413.

**W6.5 (never disable a submit to express a requirement).** `R` A missing
required value is surfaced through `aria-describedby` and an inline
message; the submit stays enabled and the server answers. The one
legitimate disabled submit is a typed confirmation on a destructive
dialog, visible and controlled by the person typing. A submit may be
disabled for in-flight state alone, with a spinner and a working label.
Src: TK `typed-confirm.ts:1-5`, ADS:917-923.

**W6.6 (hints).** `R` A hint is one snippet with a derived id, wired by
`aria-describedby` only when present, sitting outside the wrapping label
(W6.2). Where native `required` would lie (every checkbox in a closed
multiselect), the custom validity message is set by hand so the browser's
own report still fires. The hint reserves its space so the layout does
not shift when it appears. DaisyUI's `.validator-hint` paints as the
error fill and resolves through W2.7's ink instead. Src: TK
`FieldInput.svelte:130-147`.

**W6.7 (validation posture).** `R` diff-reviewer. A developer's mistake
throws at declaration time. A member's input comes back as data:
save-time validation returns a result, never throws, and accumulates.
Data a client re-posts is re-validated field by field where it crosses
the trust boundary, which for dubplate is the Go handler. Src: CI:22-27,
CI:46-50.

**W6.8 (debt versus gate).** `g` For each field the uploader or an admin
form collects, the plan says whether a missing value blocks submission or
is recorded as debt with a non-blocking notice. cairn's alt-text ruling
(debt, never a gate) is the shape; dubplate's incomplete tags on an
upload are rung 7's explicit decision. Src: ADS:809-821.

**W6.9 (hidden fields still submit).** `R` A panel toggled with the
`hidden` attribute leaves the accessibility tree and tab order while its
`display: none` fields still submit. Src: ADS:556-562.

**W6.10 (every drag affordance has a single-pointer equivalent).** `G`
rendered assertion: a page carrying a drop target carries a file input.
SC 2.5.7 requires that functionality operated by dragging be achievable
by a single pointer without dragging. The uploader's equivalent is a real
`<input type="file">` on the same surface, which the uploader design
already makes a first-class route rather than a degraded one. The same
criterion binds any later drag-to-reorder.

**W6.11 (state survives a retry, and input purpose is declared).** `R`
daisyui-a11y-reviewer. SC 3.3.7 is Level A: information the person
already entered is available on a later step of the same process. W7.15's
escape hatch back to the login form carries the address the person typed,
and a failed upload's retry never re-asks for metadata the system already
derived. Every field whose purpose is in WCAG's input-purpose list
carries the matching `autocomplete` token (SC 1.3.5); the login field's
`autocomplete="email"` is R5's instance, and this clause is its home for
the next form.

## 7. Feedback states

Unless a clause says otherwise its mark is `R`, read by
`daisyui-a11y-reviewer` at each pass gate.

**W7.1 (two empty states).** `R` A whole-concept empty state (no albums
yet) drops the list card and centers on the content area, so a first-run
screen reads as composed rather than as a small box at the top of a tall
page. A filtered-to-zero state is an in-card notice owned by the table
component, echoing the query and offering the way back, announced through
W7.6's region. The caller passes bare content and adds no size, color, or
alignment class. The empty state's heading level is a prop: `h1` when it
is the page's only content. Src: ADS:448-458.

**W7.2 (the empty state's shape).** `R` A mark or illustration,
concept-named copy, and the surface's own primary action. Not a bare line
of text. Which mark, and whether the action is an upload, an add, or
something else, is the surface's. This is one of W1.3's homes for
personality. Src: ADS:448-450.

**W7.3 (loading).** `R` A long-running submit flips a local `$state` on
submit and shows a spinner with a working label. A loading region is
`role="status"` with the spinner `aria-hidden`. Src: ADS:378-380.

**W7.4 (honest async states).** `R` An async result has three honest
states and never a fabricated finished one: a settle cue that is a visual
chip, not a live region, so it does not announce on every edit; an
incomplete skeleton naming each empty required region; and a failed
surface that keeps the form. The uploader's per-album verdict is this
shape. Src: ADS:511-516.

**W7.5 (`status` versus `alert`).** `R` `role="alert"` is for a
post-action failure, where an action was attempted and failed, and it
interrupts. `role="status"` is for everything else, including a
pre-action refusal and a success note. A region whose content never
changes is not a live region at all. Src: TK `PreviewBanner.svelte:15-18`.

**W7.6 (the live-region helper, built once).** `G` a standing proof test.
One persistent `sr-only` live region per screen, rendered unconditionally
so it can announce when it fills. The visible alert drops its `role`
entirely, since an `{#if}`-gated role element inserted fresh announces
inconsistently. Where two things must be announced (a bulk tally and the
last single action), two regions exist, because one cannot voice both
without clobbering itself. dubplate ships this as one helper module.

A repeated identical error re-announces by clearing the region and
setting the text on the next frame, because the region's text only
announces when it changes. If a nonce is used instead it is U+00A0, not a
zero-width space: the field reports settled on the no-break space, and no
vendor or W3C source states U+200B's behavior across NVDA, JAWS, and
VoiceOver. Either mechanism carries the standing proof test W11.7
requires of a tuned value, naming the screen-reader and browser pairs it
was measured on. That measurement is owed; it is not made here. Src: TK
`ConceptList.svelte:265-310`; WordPress core ticket 36853.

**W7.7 (destructive confirm).** `R` daisyui-a11y-reviewer. A modal
`<dialog role="alertdialog">` whose `aria-describedby` points at the
element carrying the alert message, without which the dialog announces
its title and nothing else. No `aria-modal` on top of the native element.
No light dismiss (no `method="dialog"` backdrop form). Escape is
suppressed only while an operation is in flight, bounded by that
operation and paired with a visible busy state, so it is never a keyboard
trap. The submit is labeled with the noun. A blocked branch refuses and
lists the blockers rather than warning. Light dismiss on the backdrop is
for non-destructive dialogs only. Src: ADS:917-923; ARIA APG.

**W7.8 (guarded, not disabled).** `G` a rendered assertion that the
reason is reachable and clears the contrast floor. A control that cannot
act right now carries `aria-disabled`, a guarded class, a resting fill at
the weight a real inactive control has, the `not-allowed` cursor, and a
reason string. The pattern ships its own override: DaisyUI applies
`pointer-events: none` to `.btn[aria-disabled="true"]`, which kills the
hover that surfaces the reason and the cursor with it, so a pinned
unlayered rule in W5.22's file restores `pointer-events: auto` and the
guarded fill and ink, version-coupled under W13.2. `aria-disabled` is
advisory, so the control stays focusable and activatable: the activation
handler no-ops and surfaces the reason through W7.6's live region, so the
reason is reachable by keyboard and not by hover alone. Native `disabled`
is for the mid-submit busy state alone. Src: ADS:546-551.

**W7.9 (never hue alone).** `R` A selected segment shows a check glyph; a
status chip's label names the state in text; a done step gets a filled
glyph and a visible tag; a verdict list colors by glyph, never hue alone.
An active or pressed state in the tint families also carries a 1px inset
hairline in its family's ink. For DaisyUI's `.btn-active` family the
hairline is dark-only: cairn measured a lightened active fill at 1.14:1
against an unselected sibling on dark, where the hairline is the whole
3:1 cue, and on light the fill separates nothing (1.06:1 to 1.31:1) so
the check glyph is the state cue, full stop. Two pairs are measured and
recorded for SC 1.4.11, the hairline against the fill it sits on and the
active state against its adjacent inactive sibling, and both route into
W12.4's `border-contrast` rule, error tier for this case alone. Src:
ADS:570-591.

**W7.10 (status chips).** `R` The chip vocabulary is closed, each
register carrying shape and color together, because a separate small tone
dot measured illegible across three screens and the two were fused. **The
set's size and membership are rung 7's**, derived from the uploader's own
journey vocabulary, which is eight states; cairn's three registers are
not inherited. An optional legend surfaces as both a native `title` and
an `sr-only` span, never `aria-label`, and is omitted where the label
explains itself; the `title` is a pointer-only enhancement, so no
information exists only there. A badge naming an identity (admin) stays a
plain badge. Src: TK `StatusChip.svelte:3-25`.

**W7.11 (safety-ranked review, for the admin decision queue).** `R`
diff-reviewer. In a review of many items, the objective items read quiet,
open pre-accepted, and are swept by one accept action. The judgment items
carry a warning-ink edge, open undecided, and are never swept until
confirmed. A judgment item carries a because-line the system can compute;
an item with no computable rationale is not offered. Nothing writes until
the accept, which is one batched transaction.

**This is rung 13's admin decision queue, not the uploader.** The
uploader addendum abolished contributor-facing review: a confident match
is accepted and the page shows what it matched as information rather than
as a gate, and a non-confident album never enters and the screen names
the failure with fix paths. The uploader's verdict is W7.4's three honest
async states plus the push-back screen. The organizer's admin-only queue
is the side door for timer-path material and is not the decision queue
either. Src: ADS:946-963; the uploader design addendum.

**W7.12 (a no-op is quiet).** `R` A run with nothing to do shows a quiet
"Nothing to fix" and never opens an empty review surface. Src:
ADS:932-937.

**W7.13 (truthful visibility).** `R` When a capability is not configured,
its controls are absent, not shown disabled, replaced by a
`role="region"` note carrying a read-only checklist of what the admin
needs to do. No control sits in the tab order. **This binds the web admin
surface at 2.0**, which is where the configuration homes spec puts it;
rung 14 lands the API those surfaces call and builds no page. It binds
rung 13's admin review queue and rung 9's repair verbs if either gets a
page before then. Src: ADS:964-973.

**W7.14 (a message states only what the system knows).** `R`
web-auth-security-reviewer. A cleared cookie jar and a second device look
alike to the server, so the message never asserts which happened. The
instruction offered is the one that works in every case. The page
`<title>` swaps per state, because it is the first thing a screen reader
announces and the one landing signal a person gets before reading. A
message panel that takes focus carries `tabindex="-1"` **and** the code
that renders it moves focus there after `tick()`; the attribute alone
does nothing, since assistive technology reaches any element by
structural navigation. Each state picks one answer, moving focus or
announcing through W7.6's region, and the comment says which. Src: TK
`LoginPage.svelte:135-154`.

**W7.15 (the login surface).** `R` web-auth-security-reviewer. **The
roster-absent answer is kind by default**, per Geoff's fork 4 ruling: it
names the roster's admin contact, and a threshold flag switches the whole
deployment to the generic outcome. The outcome copy is three authored
strings, not one: the generic outcome, the roster-absent line, and the
rate-limit line, all authoritative over any client-side `validator` hint.

The consumption interstitial renders a static confirm control on GET and
consumes nothing; only the explicit POST verifies, so a mail scanner
cannot burn the link. Three outcome states are distinguished, never
collapsed: a send failure (warning), a throttle (plain status), and an
unexpected failure (error), and a fresh result supersedes a stale one so
two alerts never stack. The sent state is a centered moment, not an
alert, with an escape hatch back to the form for a mistyped address,
carrying the address (W6.11). **Both auth pages carry `noindex,
nofollow`**, which is this standard's own addition and appears in no rung
5 source. Src: R5 fork 4 and R5:1392-1413; the `noindex` requirement is
new here.

**W7.16 (focus after an async unmount).** `R` When an async transition
unmounts the focused control, focus is moved deliberately in an `$effect`
keyed on the status discriminant plus `tick()`; otherwise it drops to
`<body>`. The upload state machine has this shape. Src: ADS:809-815.

**W7.17 (a progress indicator is real, or it is inert and voiced).** `G`
rendered assertion on the uploader. An upload's progress is not
decorative: the default is a `<progress>` element, or `role="progressbar"`
with an accessible name and `aria-valuenow`, so a screen-reader user can
query it on demand rather than re-navigating to static text. The
authoritative count also lives in adjacent text ("2 of 4 done"), voiced
through a throttled `role="status"` region, for SC 4.1.3. An ornamental
bar whose value is already voiced that way may be inert: non-focusable,
non-interactive, and a plain `<div>`. It is not a `<progress>` with
`role="presentation"`, which is not a conforming role on that element and
is ignored on anything carrying a global ARIA attribute. Src:
ADS:990-997; ARIA in HTML.

**W7.18 (the magic link is the accessible-authentication mechanism).**
`R` web-auth-security-reviewer. SC 3.3.8 forbids a cognitive function
test in the authentication step, and email link authentication (technique
G218) is a listed sufficient technique. That is what dubplate's login
discharges, and it is why R5 rules out a typed-code fallback. A
transcribed-code variant is admissible only if paste is supported, and no
credential field ever suppresses paste. Src: WCAG 2.2; R5 fork 6.

**W7.19 (the uploader's own surfaces are rung 7's and are not
standardized here).** `g` Four structures named so rung 7 does not read
their absence as an oversight: the drop target's states (idle, dragover,
invalid type, drop outside the target, the `dragleave`-on-child trap)
with W6.10's single-pointer equivalent; the file-picker route as a
first-class path, which is the iPad path; a multi-file progress surface
running over minutes on a residential uplink, which W7.3 and W7.17 do not
cover between them; and a collection picker that hides itself when one
collection is configured. Src: the uploader design.

## 8. Copy

Unless a clause says otherwise its mark is `R`, read by W12.3's prose
gate and by a human copy read at each pass gate.

**W8.1 (every string is brand prose).** `R` copy read. Button labels,
headings, empty states, hints, confirmations, and any page the Go server
emits outside the components are held to W1.2's register. Src:
ADS:1148-1164.

**W8.2 (the jargon-leak catalogue).** `R` copy read. The general rule is
developer vocabulary reaching a member. dubplate's own leak list, the
checklist a copy read runs: MusicBrainz, beets, tags, FLAC, transcode,
matcher, session, roster, epoch, collection. Two are not leaks in every
context: the push-back screen must say MusicBrainz, with a link and a
one-line how, because a member genuinely has to learn it, and a
collection is a product concept once a deployment has more than one.
"beets" and "matcher" never reach a member. Chatty filler is refused, and
so is a string implying the system has intentions about a member's files,
which is the honesty rule under cairn's personification ban and does not
forbid W1.3's motif. Src: cairn's admin UX audit, replaced by dubplate's
own list.

**W8.3 (concept nouns).** `G` ai-tells, in part. A string names the
concept ("No albums yet"), never a generic noun ("No entries"). Src:
ADS:1148-1164.

**W8.4 (plural-aware labels).** `R` A count line takes an
`{ one, many }` item label pair through one `itemNoun(count, label)`
function, never a bare string, so "1 tracks" has a single fix point. Src:
TK `format.ts:98-121`.

**W8.5 (an accessible name says what the control does).** `R`
daisyui-a11y-reviewer. "Search by album, artist, or uploader", never
"Search". The name contains the visible label's words, for SC 2.5.3, so a
voice user can say what they see. Src: `exemplar-list.md:92-94`.

**W8.6 (one separator, no em dash).** `G` em-dash-in-markup. A chip or
meta line joins its parts with one consistent separator glyph, which is
rung 7's choice. The em dash does not appear in a UI string. Src:
ADS:657-659.

**W8.7 (formatters are citizens).** `R` No screen hand-rolls date or
duration arithmetic. Every formatter takes locale and time zone as
options with neutral defaults, accepts nullish, and takes a `fallback`
defaulting to the empty string. A civil date parses at local midnight and
never routes through a time-of-day formatter. Tabular numerals on any
digit run that sits in a column or updates in place; a date never wraps.
Src: TK `format.ts:1-38`, CRAFT:73-76.

**W8.8 (failure vocabularies are typed, and carry kinds not wording).**
`G` a no-drift test. A wire-level failure set is a typed map, not
scattered strings, and client code surfaces failures as typed UI states,
never a bare `console.*`. The client's map carries failure kinds and
renders wording supplied by the Go layer, or it is generated from
`internal/reject` with a no-drift test. Either keeps one home, which is
the property CLAUDE.md's verbatim contract depends on. Src: CI:36-39.

**W8.9 (copy has a home).** `R` copy read. Single-use copy is written
where it renders; a repeated operational string is a `const` at the top
of its component. That is a reuse rule, independent of any gate. Rung 5's
string deck is the home for its three pages' strings. The reject copy's
authority is the ratified deck at `research/passb-copy-final.md`, and
shipped strings and templates are checked against it, never the reverse
(W12.3, quote-grounding). Src: the build spec's string-authority ruling.

## 9. Focus and motion

**W9.1 (one focus-ring language).** `G` focus-renders. One token triple
(outline color, offset, radius) is delivered two ways. A Tailwind
`@utility` serves markup, carrying the outline pair and never the radius,
since call sites set their own. The raw tokens serve a descendant
selector a class cannot reach. Where a tinted ground shares the ring's hue, a page-colored
halo sits under the ring. The rendered rule asserts a painted change at
every tab stop, the mechanical half of SC 2.4.7; whether the indicator is
visible enough is the coherence read's. Src: WM `tokens.css:156-181`.

**W9.2 (focus parity).** `R` svelte-reviewer, with a narrow static half.
Every hand-authored `:hover` has a sibling `:focus-visible` or
`:focus-within` in the same source. The static rule reads hand-authored
CSS only, and W5.7 keeps scoped blocks thin, so most hover states are
Tailwind `hover:` utilities the rule cannot see. The reviewer is the
enforcement, the gate is the narrow net, and W12.4's `focus-renders`
proves the ring paints. Src: AUD:76; RULES `static/focus-parity.ts:12-16`.

**W9.3 (the motion band).** `R` svelte-reviewer, with `G`
reduced-motion-honored. Every transition and animation **dubplate
authors** lands in 150 to 250ms; `transition: all` never ships, because
it re-animates every property a later edit adds. The band binds authored
CSS, not the built sheet: DaisyUI's sheet carries hundreds of durations
outside the band and one reduced-motion block, and a rule reading the
built sheet would fire on all of them, which W12.2 rules a
gate-integrity defect. Authored motion is gated behind
`prefers-reduced-motion` in one closing block per sheet, and the comment
says what survives. **The guarantee that reduced motion is honored is
rendered, not static**: a global kill switch covers the vendor rules too,
it is a deliberate unlayered `!important` entry in W5.22's file and
W5.8's pinned set, and the rendered gate verifies it under Playwright's
`reducedMotion: 'reduce'` emulation. A static rule can prove a media
block exists and nothing more, which is the failure cairn's pre-mortem
records. DL R2.8 rules that motion is not a theme axis. Src: AUD:77-78.

## 10. Code idioms for `web/src`

Section 10's mark is `R` diff-reviewer per task unless a clause names a
gate.

**W10.1 (no classes).** `G` ESLint. Production code has no classes, one
`Error` subclass and test doubles excepted. A runtime factory is a
`create*` closure over injected dependencies returning an object of named
functions; a factory whose whole surface is one function returns the
function. Src: CI:53-59.

**W10.2 (async).** `G` ESLint for the `.then` half. `async`/`await` only.
Sequential awaits for dependent work; `Promise.all` only for genuinely
independent fan-out. Latest-wins arbitration (the uploader's verdict
polling) is one extracted, tested shape, not inlined counters. A
discarded fetch body is drained. Src: CI:96-108.

**W10.3 (a module opens with its job).** A one-paragraph orientation
header names the module's job and the non-obvious rationale, including
which alternative was declined. Src: CI:68-72.

**W10.4 (a repeated idiom extracts once).** A shape copied into a second
file extracts to one home, named with its copy count. A repeated local
workaround is the loudest signal something sits at the wrong altitude.
Src: CI:151-154.

**W10.5 (chrome reads one store).** Header and footer read one
client-side session and config store and never fetch for themselves, with
a documented fallback so an accessible name never goes empty when a load
fails. Src: WM `SiteHeader.svelte:8-9`.

**W10.6 (icons are path data).** Icons are one module of SVG path data
keyed by name, each non-obvious glyph carrying a use comment, with no
icon package dependency. Inline SVG in chrome is `aria-hidden` and the
control carries the label. A fill-only line glyph is zero-area and paints
nothing; the standalone case takes a stroke. Src: WM `icons.ts:2-24`.

**W10.7 (an escape hatch carries its reason).** A `skipLibCheck`, a
`WATCH:` comment for a decision with a known future trigger, a
`throw`-unless-named build policy hook: each states why it exists.
`csrf.checkOrigin: false` is never copied from waymark; dubplate's Go
server owns CSRF and SvelteKit form actions are forbidden. Src: WM
`svelte.config.js:21-55`.

**W10.8 (deliberately not standardized).** `g` Section 14 is the single
list of what this standard refuses to decide, so a fork is not misread as
an oversight. Src: CI:174-182.

## 11. Verification practice

**W11.1 (the bar is the running binary).** `g` The polish bar every
surface matches is the real shipped app, served by `dubplate-server
serve` from the embedded bundle. A static HTML twin is not the bar. Src:
cairn `design/README.md:7-28`.

**W11.2 (mockups in real classes).** `g` A screen with no component yet
is mocked in the real utility classes inside the theme wrapper, compiled
by the real toolchain, never hand-rolled CSS against the tokens, so the
port is a transcription and nothing drifts between the approved
screenshot and the shipped component. A winning mockup is frozen as a
design-reference note beside it. A design fork is decided from rendered
compare pages annotated with measured values, never from adjectives. Src:
cairn `design/README.md:30-60`.

**W11.3 (the styleguide route).** `g` One growing route rendering every
token as a labeled swatch from a typed list, every type role at size with
its use, and every real component state: a verdict row, a reject reason
from `internal/reject`, an empty uploader, a failed login. It is the one
URL where every gate in section 12 sees the whole vocabulary. It labels
honestly what is demo and what has a real call site. **It is admin-only,
not dev-only, unless W11.8's fence exists**: under `adapter-static` every
route prerenders into the embedded bundle, so a route with no build-time
fence ships to adopters. It lands at rung 7 beside the fence. Src: WM
`(site)/styleguide/+page.svelte`.

**W11.4 (the done gate).** `g` Before a screen is called finished: the
static gate; the rendered gate in both schemes against `dubplate-server
serve`; then the coherence read. A clean audit means the vocabulary is
correct. It does not mean the screen is done, and a report says green
audit, never design-done. A suppression added to get there is named in
the report. Src: cairn `SKILL.md:88-109`.

**W11.5 (the coherence read, at rung 7).** `R` visual-verifier. A fixed
six-capture set (390 and 1440, light and dark, plus one interaction state
in both schemes), graded by a fresh-context reader that sees only the
images, never the build log or the diff, at k=3 with 2-of-3 consensus per
state. A state fails if either scheme fails any item; overall pass only
if every state passes. The reader works under four rules: never cite a distance not
measured; trace a misalignment across the whole row before naming it; one
differing property is not a system; check a sibling before failing on one
value's wrap.

**This is rung 7's uploader gate and is not the standing render-and-read
bar.** The standing `see-it-before-geoff-tests-it` bar is the lane's own
render and read of every page on the user path, in the main loop, before
any attended check; R5 states that rung 5's pages meet that bar and that
the visual-fidelity chain does not apply to them. Neither substitutes for
the other. Src: GRADE for the capture set; cairn's craft-chapter
acceptance record for k=3 and the 2-of-3 bar.

**W11.6 (screenshots are not geometry).** `g` A pixel-diff pass at a
tolerance floor is not evidence a small-footprint defect is absent: a
1.5px icon shift measured 51 differing pixels under a 120-pixel floor. A
defect whose footprint fits under the floor needs a test that measures
geometry. Committed baselines are the reviewed record of intended drift,
updated in the same commit as the change that shifted them. Src: cairn's
vertical-alignment harvest findings; ADS:1288-1290.

**W11.7 (component tests).** `R` diff-reviewer. Where an assertion is
visual, the test loads the compiled sheet and measures computed style and
geometry at two viewports, because a bare render carries no stylesheet
and an unstyled defect passes a DOM-only test. Tests query by role or
label text; `querySelector` only for structure semantics cannot express;
no test ids. A controlled primitive is tested through a small harness
component. Test names state the rule, not the mechanism. Interaction
primitives get one test per mechanic including the negative cases. A
tuned value gets a standing proof test named in the comment that depends
on it. Pure helpers use `it.each` with the case named from its data. Src:
CI:135-137.

**W11.8 (the dev fixture, fenced on both sides).** `G` a sentinel
assertion. The rendered gate and the screenshot suite need deterministic
content and a session. dubplate ships a dev backend behind a four-layer
fence, three client-side and one in Go. The client half is three layers: a build-time
`define` that dead-code eliminates the fixture from the embedded bundle,
a runtime opt-in, and a dynamic import, with the flag riding a Vite
plugin `config` hook rather than the function form of `defineConfig` and
never an exported constant, which SvelteKit does not fold across the
module boundary.

**The Go half is a build tag that excludes the session-minting handler
from the release binary**, because the session is a `__Host-` cookie
signed by the Go server's HMAC secret and the roster is a file the Go
layer writes, so no client-side fence can mint either. Roster seeding is
the CLI verb or a test fixture directory, never a second writer. The
bypass is an authentication breach if it reaches the box, so the gate
asserts it: the embedded bundle is searched for a sentinel string from
the fixture module and the leg fails on a hit, and a release build fails
if the tagged handler is reachable. R5's threat-model entry-point list
takes the amendment at rung 7. Src: cairn's dev package README.

**W11.9 (the design loop is gate-light).** `g` Tuning a value: change it,
watch the dev-server terminal, look at the page, decide. The full gate
runs once at the end. Every few iterations, check a narrow viewport and a
wide one. Src: cairn `design-your-site.md`.

**W11.10 (refuted findings are recorded).** `g` A pre-launch design
review carries a refuted section so a finding does not resurface. cairn's
failure catalogue is the pre-mortem checklist: an uncomposed mobile
header; an unpinned footer; links with no identity at rest;
browser-default focus rings; no `:active` state anywhere; instant hovers;
a reduced-motion gate a comment claimed and the code lacked; faux
italics; fonts with no metric-compatible fallback; no intrinsic image
dimensions; and page weight. The weight items are the ones an embedded
bundle with no CDN is most exposed to. Src: cairn's waymark design review
audit.

## 12. Gates

DL section 6 lands G1 through G10 at rung 5 and names the web leg's
rejoin. This section adds the gates the harvest found beyond them. Each
either gates or does not exist.

A `G` mark elsewhere in this document may name a committed test outside
the audit engine rather than a rule in W12.3 or W12.4. W2.3's, W2.5's and
W2.8's role-definition test, W7.6's standing proof test, W7.17's rendered
assertion on the uploader, W8.8's no-drift test, W11.8's sentinel
assertion, W13.3's inventory diff, and W13.5's baselines and geometry
assertions are those, each a test the repo runs in its own suite.

**W12.1 (refuse rather than report a clean run).** A run that could not
start or finish exits 2 and is never a design verdict. Every shape of
silent green fails: no rules registered, no pages configured, the server
not answering, the browser absent, any page outside 2xx, the redirect
trap, and a scope path the author wrote that does not exist. Three shapes
are dubplate's own, because it is a client-rendered bundle served by a Go
binary.

- **The un-booted page.** A page returning 200 with a valid shell that
  never boots, because a CSP hash does not match, a bundle chunk 404s, or
  hydration throws; every rule then measures an empty body and finds
  nothing. Each configured page declares an expected identity (a heading
  or landmark signature) in the audit config, and a page that settles
  without it fails at exit 2. This is dubplate's recorded failure mode:
  DL revision 1's CSP would have loaded rung 5's login page and never
  booted it.
- **The stale embedded bundle.** `go:embed` bakes `web/dist` at Go build
  time, so a rendered run against an older binary measures yesterday's
  paint and reports clean, invisibly, because the hashed asset it loads
  exists and is simply old. The web leg builds `web/`, rebuilds the
  binary, and refuses at exit 2 if the bundle hash served by the process
  under audit differs from the one just built.
- **SSR-versus-hydrated identity has no analogue.** cairn's engine
  compares a JavaScript-disabled capture against the settled one, and a
  client-rendered app has no server-rendered side, so that guard is
  replaced by the expected-identity assertion above, not ported.

A rule that cannot measure reports an advisory naming what it could not
read; a check that skips itself is the failure the gate exists to rule
out. The login path is configuration, not a constant, and until W11.8's
fixture exists an unauthenticated rendered run can audit only the auth
pages. Src: AUD:36-50, 141-143, 167-221.

**W12.2 (two tiers, co-located suppression, stale entries fail).** Error
tier gates; advisory tier reports and never changes the exit code,
because each advisory rule measures a compositional question a novel
component may answer differently on purpose. A finding is suppressed at
its source line by a comment naming the rule and a reason; a directive
that silences nothing is itself an error, and neither error can be
suppressed. A rendered finding with no source line uses a
page-plus-selector allowlist, and a stale entry is an error. The summary
always prints the suppression count, including zero. A rule names shapes
by tag or ARIA role, never by class, wherever the shape has one. A
false-positive rate is a gate-integrity defect and is fixed at the rule.
The
suppression directive marker carries dubplate's own name, chosen once;
adopting cairn's would brand every suppression permanently and make a
later switch a rewrite of every directive. Src: AUD:38-41, 88-113, 244.

**W12.3 (static rules).** Parsed from `.svelte` files under `web/src` and
from the built sheet, no browser. Classes written from TypeScript are
outside the scan, a known limit rather than a silent one. Nine rules are
stack-generic and arrive with the engine on either path of W12.8; the
rest are dubplate's own build, because cairn's four prose and idiom
checks (`check-admin-prose.mjs`, `check-editor-quotes.mjs`,
`check-idioms.mjs`, `check-public-tokens.mjs`) live under `scripts/`,
which is not in the published package's `files` list.

| Rule | Asserts | Origin |
|---|---|---|
| uncompiled-class | Every class a component writes compiles into the built sheet or is declared in its own scoped style | engine |
| achromatic | No achromatic literal (a hex, `rgb()`, or `hsl()` grey, black, or white) is declared where a palette name belongs. It extends DL G1 into the palette files, which DL R2.4 and G1 leave as this rule's only possible subject | engine |
| motion-band, reduced-motion | W9.3, over authored CSS only | engine |
| focus-parity | W9.2's narrow source half | engine |
| refused-defaults | W5.17's table, keyed on the open mechanism where there is one and the class name otherwise; four of six rows out of the box, extended by W5.23's raw-class check | engine, extended |
| unlayered-font-clobber | W5.8's scoped-style clause | engine |
| list-role | W5.16's affirmative form, item-class path | engine |
| stripe-trim-parity | A zebra stripe and an unconditioned first/last-child trim never share a row class | engine |
| unlayered-count | W5.8's pinned set by whole-selector equality, a ratchet raised only by the commit that adds rules | dubplate |
| role-utility | W2.4, replacing cairn's `gap-scale` | dubplate |
| spacing-collision | W2.10's first trap against the pinned Tailwind's `--container-*` set | dubplate |
| element-id, em-dash-in-markup, aria-sort-placement | W5.19, W8.6, W5.15, each a regex over markup | dubplate |
| ai-tells | Every string extracted from `web/src` markup and literals passes the mechanical AI and marketing tells, with a `--list` mode for a release-time human read. **It enforces no register** | dubplate, ported |
| quote-grounding | Every reject reason shipped in source, a doc, an email template, or `upload-check` is grounded in the ratified copy deck, checked deck-first, each literal segment appearing in order with a minimum length so a wildcard cannot vacuously match | dubplate, ported |
| no-process-citation | No `Task N`, `Phase N`, `this pass`, or `docs/superpowers/` path in shipped source | dubplate, ported |
| contrast-extension | Extends DL G9's subject to every text-bearing pair, W2.7's measured inks included, measured in sRGB per DL R2.6 | dubplate, ported |
| version-stamp, pin-exactness | W13.2 and W13.7 | dubplate |

A display-p3 leg is not gated here. DL R2.6 fixes the method, ratios
computed on sRGB relative luminance per WCAG 2.2, so a wide-gamut leg
would change a ratified gate's measurement without evidence. Whether a
wide-gamut display resolves an `oklch()` value differently enough to fail
a pair DL G9 passes is a rung 7 evaluation, and a p3 leg is added only if
that evaluation produces a failing pair, with the erratum DL is then owed.

The contrast-extension rule does **not** re-run with the operator overlay
merged. DL R2.6 puts tiers 2 and 3 outside the floor, and the branding stylesheet
is generated from config at serve time on the adopter's box, so no
build-time gate has it. A serve-time `doctor` advisory over a
deployment's own rendered branding is the honest form. `token-resolution`
and `gap-scale` are dropped: DL G2 already asserts each theme file's
exact declared name set, and `gap-scale`'s subject is the bracket
utilities W2.4 forbids outright.

**W12.4 (rendered rules).** Driven by Playwright against `dubplate-server
serve` already answering, both schemes, at rest and with a menu or dialog
open. The audited artifact is the exact embedded bundle the box ships,
which W12.1 makes a checked property rather than a claim.

| Rule | Tier | Asserts |
|---|---|---|
| viewport-overflow | error | At 390 and 320 nothing paints past the viewport, including an element wider than its box; gated on the document actually scrolling; deliberate scroll containers skipped |
| touch-targets | error | W4.10's square test on the activation region, at every W4.1 viewport, carrying SC 2.5.8's five exceptions |
| interactive-contrast | error | An interactive element's text against what actually paints behind it, by canvas readback |
| focus-renders | error | Tabbing to each stop produces a painted change, not merely a selector in the sheet |
| one-filled-action | error | At most one accent-filled control per surface, topmost open layer, partitioned by landmarks, read from the computed background |
| list-role | error | W5.16, descendant-selector path |
| h1-count | error | One `h1` per page. dubplate's own small rule, replacing cairn's `screen-anatomy` |
| skip-link, contents-wrapper, reduced-motion-honored, file-input-present, guarded-reachable | error | W3.6, W4.8, W9.3, W6.10, W7.8, one assertion each |
| border-contrast | error for W7.9's two pairs, advisory otherwise | Borders and hairlines at 3:1 |
| chip-ground-collision | advisory | A badge's painted fill is not camouflaged against its row, 1.5:1 |
| field-edge-alignment | advisory | Inputs in one column share a left edge within 1.5px |
| container-inset-asymmetry | advisory | Content sits as far from a container's left edge as its right, ported off cairn's selector list |
| form-font-parity | advisory | Every form control computes the theme root's family |
| weight-budget | advisory | Two weights per content region |
| text-spacing | advisory | Nothing clips under SC 1.4.12's 1.5x line-height and 0.12em letter-spacing, which W3.4's pinned height and W4.10's fixed-height boxes risk |

The advisory tier follows cairn's own rulings, one recorded after a
measured 60% false-positive rate, and W12.2 forbids promoting a heuristic
without its own re-measured evidence.

Four of cairn's sixteen rendered rules are not used. `norms-bands` grades
a tree against cairn's measured control geometry, which owner ruling 1
forbids. `panel-width` measures a cairn component dubplate does not ship.
`screen-anatomy` is replaced by `h1-count`. `relational-spacing` is
deferred past rung 7: advisory, expensive, and welded to cairn's token
names. Six rules measure nothing until rung 7 builds the surface
(`stripe-trim-parity`, `chip-ground-collision`, `one-filled-action`,
`field-edge-alignment`, `weight-budget`, `container-inset-asymmetry`),
and section 15 routes them there rather than landing them empty at 5c.

**W12.5 (colors are read back, never parsed).** Every color comparison
paints the computed color on a canvas and reads sRGB bytes back. cairn's
earlier parser could not read one oklch value and had been passing
vacuously; a parser is the one component guaranteed to be wrong about a
real value, and the first thing a new color syntax breaks. Src:
AUD:252-255.

**W12.6 (where each leg runs).** The static rules are cheap and they are
where the typo class lives, so they run in `scripts/check.sh`'s web leg
on every commit.

**The rendered leg does not run per task.** It runs at the pass-end gate,
in CI on every push, and in a per-task gate only for a task that moves
rendered paint or touches `web/`. That is CLAUDE.md's gate-economy ruling
(Geoff, 2026-09-09) applied rather than excepted: a browser suite joins
the per-task gate only where it can catch something, and a task in
`internal/engine/tagio` moves no paint. The leg fails loudly rather than
skipping when a diff did touch `web/`, which is W12.1's doctrine held.

Three operational facts. The browser is a hard-required tool, named and
failing by name, the way rung 5c makes Node a hard requirement under DL
R1.5 and DL section 6's rule that a check either gates or does not exist.
Playwright drives its own downloaded Chromium rather than the
workstation's layered `chromium-browser`. The server port is an environment variable, because
the lane model runs two worktrees concurrently. The Chromium run is
contained under `systemd-run --user --scope` with a memory cap, falling
back to unconstrained where no user bus exists.

Measured cost is in the reference. `bash scripts/check.sh` is about two
minutes today, and a bare Playwright floor is 9.2 seconds, 12.8 including
node start. The real leg's total is unmeasured until rung 7 builds it,
since it adds per-stop focus comparison, per-element canvas readbacks,
and two viewport resizes per page, and that is the second reason it does
not sit in the per-commit gate.

**W12.7 (toolchain).** `svelte-check` with strict TypeScript, prettier at
100 columns with the Svelte plugin and double quotes in CSS, vitest with
a `unit` project in Node and a `component` project in real Chromium
through `vitest-browser-svelte`, and ESLint carrying the `ts-conventions`
jsdoc and tsdoc config, a `no-em-dash-in-comments` rule, and the
no-classes and no-`.then` rules W10.1 and W10.2 name. Path aliases are
declared in the SvelteKit config and the vitest config both, with the
reason in a comment. The `component` project sits in the rendered leg's
scope under W12.6, not the per-commit one.

**W12.8 (port, and where a dependency is even possible).** **Ruled,
conductor 2026-09-12: a port is the primary path, and the dependency is
not evaluated for the rendered rules at all.**

The rendered half cannot be a dependency. cairn's runner establishes page
identity by comparing a JavaScript-disabled capture against the settled
one, on title and landmark. A client-rendered bundle has no
server-rendered side, so the identities never match, every page raises an
error-tier `page-identity-mismatch`, and zero rules run; configure the
login page alongside another and the redirect trap fires unconditionally
at exit 2 instead. Neither outcome is an audit, and the defect is
structural rather than configuration.

The static half could be a dependency for the nine generic rules, but the
published CLI cannot subset its registry, so a dependency drags in the
rules welded to cairn's token names, classes, and routes (ten rules and
three harness constants, enumerated in the reference) along with a
35-dependency CMS to obtain one bin. Whether to install it or to lift
nine small MIT-licensed files is rung 7's plan decision, judged on the
2.0 spec's adoption bar with the evidence recorded there rather than
asserted here. The license leg is clear: cairn is MIT.

**W12.9 (the conformance target).** WCAG 2.2 Level AA for every surface,
with SC 2.5.5 at AAA held as the 44px design floor (W4.10). The criteria
this standard discharges by mechanism: 1.3.5 and 3.3.7 (W6.11), 1.4.3
(W2.7), 1.4.4 (W4.9), 1.4.10 (W4.5), 1.4.11 (W5.18, W7.9), 1.4.12
(W12.4's `text-spacing`), 1.4.13 (W7.10's pointer-only `title`), 2.1.2
(W7.7's bounded Escape suppression), 2.2.1 (R5's link expiry and session
cap, under the criterion's essential exception, which covers a
security-bounded session), 2.4.1 and 2.4.11 (W3.6),
2.4.7 (W9.1, W9.2), 2.5.3 (W8.5), 2.5.5 and 2.5.8 (W4.10), 2.5.7
(W6.10), 3.3.8 (W7.18), and 4.1.3 (W7.3, W7.5, W7.6, W7.17). A new clause
names the criterion it discharges or says that it discharges none.

## 13. Upgrading the UI libraries

Owner direction, Geoff, 2026-09-12: the design makes it easy to update
these libraries over time. What this section buys is a bounded *version*
upgrade, which is not portability.

**W13.1 (what is and is not stable across a major).** `g` Neither
DaisyUI's variable names nor its class names are stable across a major:
DaisyUI 5 renamed every theme variable, removed three, introduced three,
and moved configuration into the CSS `@plugin` directive, while
comparatively few component classes went away. The variables are the
smaller and better-documented blast radius, not the safe one. What still
follows: a gate rule names a shape by tag or ARIA role wherever the shape
has one (W12.2); components reach DaisyUI through documented classes and
modifiers rather than internal selectors; every internal-selector
override is quarantined in one file (W5.22); every component's comment
names the classes it assembles from (W5.7); and the coupled families are
reached through recipes (W5.23). Together those make the blast radius a
grep and a directory.

**W13.2 (version-coupled values have one home and a gate).** `G`
version-stamp. Some clauses carry a value that is a version's answer to a
requirement rather than the requirement itself. Each lives in one
committed module or table carrying the value, the library and exact
version it was measured against, the clause it serves, and the
measurement method. A prose-only list goes stale, which is the failure
W4.2 names.

**The stamp is a re-measurement obligation, not an equality gate across
entries.** Entries legitimately carry different versions, because each
records the version its own measurement was taken against, and today's
list already spans more than one. What the gate asserts is that every
entry's stamp is present and well formed and that the bump rehearsal
(W13.4) re-measures each entry whose stamp is behind the pin in
`web/package.json`, recording the new value and the new stamp in the same
commit. Adopting a pin therefore carries re-measuring the entries behind
it as work, which is part of what rung 7's pin choice (W13.7) buys.

Today's entries: DL R3.7's `--size-field`; W5.18's 55% edge mix, the 20%
mix it corrects, and `.toggle`'s measured pass; each of W5.17's six rows;
W7.8's `pointer-events` override; W5.13's `.menu` `display: flex` and the
CSS-focus dropdown path; W5.16's `.list` flex and 14px; W2.10's three
Tailwind traps; W2.7's measured `--color-warning` ratio and the premise
that DaisyUI tunes its status colors as fills; and DL R2.1's variable
count, measured against `daisyui@5.7.17`.
That last is more than version-sensitive, because it is a completeness
rule: a DaisyUI 6 adding a role variable makes every dubplate theme
incomplete and DL G2 would enforce the stale number against the correct
fix. That is owed to DL (section 19).

**W13.3 (the classes this tree writes are watched, with their
declarations).** `G` inventory diff, per bump. The watched set is the
classes dubplate's own markup writes (the union of W5.7's per-component
lists) plus the role utilities, not the whole emitted sheet: cairn
watches its whole sheet because that sheet is a public API for consumer
markup, and dubplate's ships inside a binary. The diff compares
declarations, not only presence, because the most expensive upgrade
defect is a class that survives and means something else. cairn's
Tailwind 3-to-4 migration left `focus-visible:outline-none` in shipped
code across the migration and several releases after it, where it
compiles and means `outline-style: none`,
so a control had no visible focus ring and no gate could see it. Paired
with it is a rendered assertion on the properties carrying accessibility
meaning (outline width, border color, pointer-events, min-height). Src:
cairn's admin sweep.

**W13.4 (the bump rehearsal).** `g` A bump runs six steps on one branch:
bump the pin; rebuild; run the screenshot baselines in both schemes; run
the static and rendered gates; confirm every role utility and grammar
token still resolves; read the screenshot diff and commit the accepted
baselines with the reason.

**Steps 2, 4, and 5 are assertion gates and must be green. Steps 3 and 6
are a review, not a gate.** A legitimate bump that moves paint makes the
baselines red by construction, and treating a baseline diff and an
assertion failure as one event either blocks every honest bump or teaches
a reader to override a red gate. cairn's only executed bump was three
commits plus a separate baseline regeneration, and it was a minor bump
that still moved paint. So the requirement is that the bump and its
accepted baselines land on the same branch with the reason in the
message, not that they are one commit.

**A new major is noticed by a mechanism, not a sentence.** The mechanism
is a `renovate` or `dependabot` config in this repo restricted to major
bumps of the four UI packages, so the opened PR is itself the rehearsal
ticket. One committed file, workstation and CI only, never a box process,
landing with rung 7. Src: ADS:1266-1293; cairn `HISTORY.md:612-620`.

**W13.5 (the styleguide is the bump's assertion surface).** `G` baselines
plus geometry assertions. W11.3's route renders every token and every
component state, so it is the named home of the committed baseline set:
one capture per component state per scheme at two widths, regenerated
only in the bump commit with the reason. Beside it sits a geometry
assertion set on the same route (computed min-height, border-color,
outline-width, font-family on one instance of each control family),
because W11.6 records that a defect fitting under the pixel floor needs a
measurement rather than a picture. A rendered pass exercises; only an
assertion catches.

**W13.6 (deprecations are tracked, not discovered).** `g` The four
authoring practices this standard fixes (runes-only Svelte 5, no
`createEventDispatcher`, snippets over slots, no server-only SvelteKit
files) keep the tree off Svelte 5's own deprecated authoring paths. That
is what they buy and no more. No Svelte 6 has been announced, and the
breakage a SvelteKit 3 brings lands elsewhere.

**The SvelteKit 3 specifics below come from the Svelte team's September
2026 release notes, they are RC-stage claims, and each carries a `WATCH:`
comment so rung 7 re-reads the migration guide before pinning.** The
surfaces a major actually moves, each on the watch list: the adapter and
its Vite plugin seam (per those notes the RC reorganizes adapter plugins
into `pre` and `post` groups, and dubplate's whole delivery is
`adapter-static` plus a Vite build feeding `go:embed`, with W11.8 pinning
a plugin `config` hook); the config shape and the route-parameter helper
those notes name `defineParams`; route exclusion of files
named `test`, `spec`, or `stories`, which interacts with W12.7's
`component` project; path aliases, declared twice; `$app/*` module paths;
and the prerendering behavior W3.7 depends on. `svelte-check` tracks the
authoring half and sees none of the adapter half. A known future change
carries a `WATCH:` comment naming the upstream issue (W10.7).

**W13.7 (exact pins, one place).** `G` pin-exactness. Every UI library is
pinned to an exact version in `web/package.json`, gated by a static check
that the four UI dependencies carry no range specifier. No clause in this
standard or DL states a version as current except as the label on a
measured value (W13.2). Rung 7 chooses the pins and re-reads DL's
contract against them.

**W13.8 (the upgrade path for an adopter is a binary).** `g` An adopter
never runs the bump; they receive a new binary whose embedded bundle was
built and rehearsed on the workstation, which is DL R1.5's universality
rule applied to upgrades. One thing the binary does not carry is the
adopter's own tier 3 CSS file, which DL R4.4 gives the last word and puts
outside the contract, so a release whose bundle crossed a DaisyUI or
Tailwind major says so in its release note and in the binary's upgrade
output, naming that file as the thing to re-check. That is a string, and
it is the whole remedy available under R1.5.

**W13.9 (DaisyUI is load-bearing and is not swappable).** `g` dubplate is
coupled to DaisyUI at the design level, and a different component library
is a redesign, not a migration: DL R2.1's 32-name contract is 29 DaisyUI
names, DL R3.7's touch floor rides `--size-field`, W2.7 exists because of
how DaisyUI tunes three colors, and W5.13, W5.16, W5.17, W5.18, W5.22,
and W5.23 each name DaisyUI internals. **A DaisyUI major is therefore not
a `web/` event**: DL R4.3 has the Go binary generate a branding
stylesheet from config containing only names from DL R2.1's set, so a
rename breaks Go code, the config key registry, and every adopter's
stored branding values at once. A DaisyUI major carries a migration
obligation, a config-schema version or a startup validation that every
stored branding name is one the binary knows, refusing loudly rather than
painting unbranded. That is owed to `internal/config` and is named here
so rung 7 does not discover it.

## 14. What this standard does not decide

The single list. W10.8 points here.

- The visual values: every color, size, face, radius, and motif. Rung 7.
- The type-role count, the gap-role numbers, the dialog cap, the
  separator glyph, the submission row's alignment, and W4.4's stacked
  gap. Rung 7.
- Which context each surface belongs to (W3.1), and whether the uploader
  is a desk.
- Which upload fields are debt and which are gates (W6.8).
- The chip register set's size and membership (W7.10).
- The uploader's four direct-manipulation structures (W7.19).
- The UI library pins (W13.7).
- The rendered touch floor's error tier for 44px. Geoff, under W4.10.
- Whether the nine generic static rules arrive as a dependency or a lift.
  Rung 7's plan, under W12.8.
- W7.6's nonce mechanism, pending its measurement.
- Forced-colors posture, neither required nor declined here. Playwright's
  `forcedColors: 'active'` makes it a cheap seventh capture if rung 7
  wants one.

## 15. Routing and the 2.0 hook statement

**Rung 5c lands only what its ratified deliverable list already carries,
and its 4.5M ceiling is unchanged by this standard.** That list is the
three pages, the composed theme as data, the theme binding, the fonts and
embed path, the R4.5 rewrite, the branding stylesheet and operator file,
the served header set, the `member` verb group, and the gate leg with
DL's G1 through G10.

**No new artifact moves into 5c.** The grammar-token module and its
inventory (W2.1 through W2.8), W2.7's measured ink split, W5.10's theme
pick, W7.6's live-region helper, the styleguide (W11.3), the dev fixture
(W11.8), and every W12 rule beyond the AI-tells one are new artifacts
rather than properties of the three pages, and adding them to 5c would
roughly double a pass whose deliverable count and ceiling were derived
item by item.

**Every clause that is a property of a page as built still binds 5c**,
sections 4, 5, 7, and 10 included. W5's preamble binds every `.svelte`
file under `web/src` and 5c is where `web/src` is born; section 10's
preamble scopes its idioms the same way. So W4.1's five-viewport bar,
W4.10's two touch floors, W4.4's and W4.6's stacking, W5.1 through W5.7
and the rest of W5's contracts, W7.1's and W7.2's empty states on the
add-member roster list, W8.4's plural-aware count, and W10.1 through
W10.7 all bind the three pages. Binding alongside them are W1.2 and
W1.4's register, W3.3's single-purpose-form anatomy and its list-screen
header, W3.6, W3.7's 404 split, W6.1 through W6.6 and
W6.11, W7.14, W7.15, W7.18, W8.1 through W8.9, and W9.1. Where a clause's
own gate lands at rung 7, the reviewer reads the clause inside 5c
instead; `svelte-reviewer` and `daisyui-a11y-reviewer` carry this
document as their lens in that pass.

**This leaves the 5c ceiling where it was**, because a property is not a
deliverable. Binding W4.1 to the login page adds no artifact to 5c's
list; it says what the login page already had to be, and the cost is a
reviewer's read and a lane's own render at the widths it was going to
render at anyway.

The AI-tells rule also runs inside 5c. **It is additional, not a register
gate**: W12.3's row says in terms that it enforces no register, and R5's
own register gate on the string deck (the Microsoft editor standard,
R5:1429-1435) applies in the form R5 states and is a separate obligation
5c carries.

One consequence is stated here rather than discovered later. **Rung 5's
three pages are authored with Tailwind's own steps where the grammar
layer does not yet exist.** W2.4 forbids a bare step only for a role the
grammar covers, and at 5c it covers none, so there is no conflict. Rung 7
lands the grammar layer and converts three pages, which is cheaper than
growing 5c.

**Rung 7** lands the rest: the component set with the disclosure and
dialog primitives (W5), the override file and control recipes (W5.22,
W5.23), the reusable table recipe and ledger-row recipe (W4.5, W4.6),
whose stacking requirement 5c's roster list already meets by hand, the
verdict model (W7.4) and the uploader's own surfaces (W7.19, W6.10), the grammar layer,
the dev fixture, the styleguide and its baselines (W11.3, W13.5), the
coherence read as the pass gate (W11.5), the bump watcher (W13.4), and
the web leg's static and rendered rules (W12), the port-or-lift decision
included.

**Rung 13** takes W7.11's safety-ranked review. **Rung 14** lands the API
the admin surfaces call and takes no clause from this standard. **2.0's
web admin surface** takes W7.13.

**2.0 hooks.** This standard lands no new hook and touches one existing
one: hook 5, the stable identity, which DL R4.2 already lands through the
per-member theme row; W5.10 changes nothing about its key, since the
signed-out cookie carries no identity. Every other
hook in S20's list is neutral to *this* standard, which is narrower than
saying a presentation standard cannot land one: DL itself lands hook 6 by
declaring theming out of the driver seam. The 2.1 first-party web player
is bound by every clause here, which is why W4.1 binds every surface
rather than an enumerated set.

## 16. Rulings Geoff reads first

1. W1.2's register line, W1.3's three personality homes with their worked
   examples, and the rule that a personality choice outranks a harvested
   mechanic. The wording is the ask.
2. W4.10's two-floor reconciliation and the conductor's tiering.
3. W12.8's port-first ruling and W12.6's removal of the rendered leg from
   the per-commit gate, which together decide what rung 7's plan builds.
4. W7.11's retargeting at rung 13 and W7.13's at 2.0, which move two
   surfaces out of the ladder's middle.
5. W5.10's cookie, which supersedes a conductor ruling in DL R4.2 rather
   than filling a gap, and section 19's row recording it.

## 19. Supersessions and errata

Sections 17 and 18 moved to the reference and their numbers are retired.
This section records what this standard changes in a ratified document,
one line each. Each is owed as an erratum in the document named, not
asserted here.

| Ratified passage | What this standard asks of it | Clause |
|---|---|---|
| CLAUDE.md, the engine paragraph naming `text-box-trim` as a silent default | The ruling was measured and not landed, and the feature is not Baseline; the example is stale | reference |
| CLAUDE.md and `scripts/check.sh`, the web leg's contents | The leg gains the static rules per commit and the rendered rules at pass end and on `web/` changes | W12.6 |
| DL R2.1 and G2, "exactly these 32 names" | A completeness rule cannot survive a DaisyUI major that adds a role variable; DL owes the version-coupled treatment | W13.2 |
| DL R4.2, the ruled `localStorage` pick | The signed-out pick is a cookie the Go layer reads while it rewrites the document; `localStorage` is not used for the pick. The reason is that a `localStorage` read needs an inline script the CSP would block | W5.10 |
| DL R4.5, "the embedded `index.html`" | Prerendering emits one document per owned route; the one documented rewrite and DL R5.6's hash set cover each of them, not a single file | W3.7 |
| DL R4.3 and `internal/config` | A DaisyUI major renaming the palette set breaks the Go branding renderer, the key registry, and stored adopter values; a migration obligation is owed | W13.9 |
| R5's threat model, entry points | W11.8's dev fixture adds a Go-side bypass behind a build tag; the entry-point list takes the amendment at rung 7 | W11.8 |
| The build spec's ladder entry for rung 14 | Rung 14 lands the admin API and no page; W7.13 binds 2.0's web admin surface | W7.13 |
