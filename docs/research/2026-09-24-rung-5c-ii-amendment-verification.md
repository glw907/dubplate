# Rung 5c-ii plan revision 2.3: amendment verification

A fresh-context read of the visual identity amendment to
`docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`
(revision 2.3 at `30f2d5b`, compared with revision 2.2 at `HEAD~1`). The
amendment record is the "Amendment for the visual identity ruling" section
of `2026-09-24-rung-5c-ii-plan-fold-record.md`. The mechanism holds: a
scratch build at the pins, rendered in headless Chromium, follows the
emulated preference with the right background and `color-scheme`. One
claim carried over from revision 2.2 is false for two themes. As written,
Task 4's acceptance row would fail a correct implementation (A1). Counts:
0 blocker, 1 major, 3 minor, 7 low.

## How each mechanism was checked

Probes ran in `/tmp/claude-1000/.../scratchpad/probe/`. That directory held
`daisyui@5.7.45` and `tailwindcss@4.3.3` with `@tailwindcss/cli@4.3.3`. It
also held the plan's file layout, abbreviated to a few variables per block,
and `@plugin "daisyui" { themes: light --default, dark; }` as Task 3 writes
it. The build used the CLI, not `@tailwindcss/vite`, but the optimizer is
the same Lightning CSS pass.

| Claim | Source | Result |
| --- | --- | --- |
| `prefersdark` emits only `:root:not([data-theme])` under the media block | `theme/index.js` | Quoted exactly |
| The plugin emits `color-scheme` with `?? "normal"` on every block | `theme/index.js` | Quoted exactly |
| The root background rule | `base/rootcolor.css` | Quoted exactly. The file also holds `:root{background-color:var(--page-scroll-bg,var(--root-bg))}`, which does no harm |
| Tailwind's `dark` variant is media-based | `dist/lib.js` | Quoted exactly |
| `system.css` and `inks.css` land in `@layer base` | Build | Yes. They share the `base` layer with DaisyUI's theme blocks, so the unlayered `branding.css` outranks them |
| The bundle's only `prefers-color-scheme` block is the system rule, and it has no `:root:not([data-theme])` | Build | Yes. `light --default` emits `:where(:root)` at (0,0,0), which the (0,1,0) system rule beats |
| Two same-named theme blocks "build to one merged rule holding both files' variables" | Build | **False for two themes (A1)** |
| `dubplate-system` follows the preference | Chromium, `--blink-settings=preferredColorScheme=0/1` | Dark gives `rgb(28,25,22)` with `color-scheme: dark`. Light gives `rgb(244,234,211)` with `color-scheme: light` |
| A single-scheme meta pins `prefers-color-scheme` | CSS Color Adjustment 1 ED; Chromium | The spec text is quoted accurately. Chromium does not pin it on a top-level document (A4) |
| MQ5 section 12.5, value `light \| dark`, the negation advice | drafts.csswg.org | Matches. The spec defines no `no-preference` value, so the plan's "no preference" case is the UA default of light |
| `Request.Cookie` doc, `readCookies` first-parse, the quote strip, and the byte set | go1.27.1 `net/http/cookie.go`, `request.go` | Quoted exactly |
| The four marker ratios and the pass figures | WCAG 2.2 luminance recomputed | All reproduce: 1.83, 1.83, 3.18, 1.09; `-content` minimum 4.96; base-content minimum 13.87; inks 6.02 and 6.34 |

## Findings, ranked by consequence

### A1 (Major): Task 4's "one merged rule" claim is false, so its bundle row fails a correct build

**Location.** Plan `:2092-2103` (the binding) and `:2150-2156` (the
acceptance row: "holds a rule whose selector list carries
`[data-theme=<name>]` ... and declares all 32 names", and "`color-scheme`
is `light` ... never `normal`").

**Defect.** Revision 2.2's measurement held for one theme, because the
match and palette rules were adjacent with the same selector, so the
minifier merged them. With the amendment's import order (`dubplate.css`
holding both match blocks, then `label.css`, then `label-night.css`), the
emitted order is match-label, match-night, palette-label, palette-night.
The probe's minified output has three rules:

- one merged match rule whose selector list carries *both* names and
  declares `color-scheme:normal`
- one palette rule per theme

The unminified output has four rules. No single rule declares all 32
names, so the row as written fails the correct implementation. A reader
of the match rule alone sees `normal`. The cascade result is still right,
because each palette rule comes later at equal specificity and in the
same layer. Interleaving per-theme match files yields one merged rule per
theme, but that rule declares `color-scheme` twice (`normal`, then
`light`). A per-rule reader is wrong either way.

**Proposed fold.**

1. Drop the "one merged rule" sentence.
2. Define the row on the cascade-effective value: for each name, the last
   declaration in source order across every rule outside `@media` whose
   selector list carries the compound `[data-theme=<name>]`. The row
   requires all 32 names to be effective, and `color-scheme` to be
   effectively `light` or `dark`.

Mutation 7 still dies under this reading (probed: with `label.css`
imported first, the effective `color-scheme` is `normal`). Cite the probe
rather than "measured".

### A2 (Minor): the bundle row has no named gate home

**Location.** Task 4, Files (`:2061-2069`), and the acceptance row at
`:2150`.

**Defect.** Task 4 creates only `g2.ts` and `g9.ts`, and R-5cii15 says G2
reads the source data files. The bundle row checks four things:

- the import order (mutation 7's only detector)
- the lone `prefers-color-scheme` block
- no `:root:not([data-theme])`
- the stray `dark:` guard that "5c-iii's pages inherit"

A one-time check at Task 4 would satisfy all four. After that, an
`app.css` reorder, or a `dark:` class in 5c-iii, regresses silently.

**Proposed fold.** Name the file and its leg: either a bundle check under
`web/scripts/gates/` run after the build in the `web` target, or a Go
test beside Task 5's embed test. State that it is gate-resident.

### A3 (Minor): G2's system parity can pass on a deleted declaration

**Location.** Task 4 states (`:2124-2126`) and mutation 4 (`:2192`).

**Defect.** The parity row fires on "a branch whose value differs". The
"missing one name" row speaks of "a block", which reads as a
`daisyui/theme` block. A G2 that iterates the branch's own declarations
passes when `--color-primary` is deleted from the dark branch. Under
`dubplate-system` on a dark device, that name would then take the light
branch's value, which is exactly the drift the copy risks. Mutation 4
tests a name-only comparison, not a branch-driven iteration. A repeated
declaration inside a branch, where CSS takes the last and a parser may
take the first, is also unstated.

**Proposed fold.**

- Add a state row: a `system.css` branch missing a name from, or
  carrying a name beyond, the 32 names plus the ink set fails, naming the
  branch and the name. A name declared twice in one branch fails.
- Add a mutation: iterate only the branch's own declarations, so a
  deleted dark-branch `--color-primary` passes.

A2's gate-resident bundle row would back this up at the built level.

### A4 (Minor, for the spec author or Geoff): branding now flattens both palettes

**Location.** R-5cii14 (`:1084`, one `:root` block) and R-5cii22, which
does not address it.

**Defect.** Before the amendment, `DUBPLATE_BRANDING`'s single `:root`
block overrode one palette. It now overrides `label`, `label-night`, and
both `dubplate-system` branches with one value. For example,
`--color-base-100` set for a light brand also paints a dark device's
ground. G8 checks names and grammar, not contrast, so an adopter's single
brand color can fail R2.6 in one mode with no gate noticing. The
amendment decided this silently, by leaving R-5cii14 unchanged.

**Proposed fold.** Record it in Task 8's handoff as a visual identity
spec question. The options are a per-scheme branding block, a note that
branding is single-scheme, or G9 over the branded result. It need not
block 5c-ii.

### A5 (Low): the meta rationale overstates the browser behavior

**Location.** `:1320-1324`.

**Defect.** "A single-scheme meta would pin the media query and break the
system branch" is the CSS Color Adjustment 1 editor's draft's model.
Chromium does not implement it for a top-level document: with
`content="light"` and a dark preference, the probe reported
`matchMedia(dark)` as true. The decision to keep the meta at
`light dark` is sound under either model.

**Proposed fold.** Qualify the sentence to cite the spec's model and note
that Chromium diverges. Optionally, route to 5c-iii's first-paint proof
whether a pinned variant (`label` or `label-night`) should carry a
single-scheme meta, which matters only for the pre-stylesheet canvas.

### A6 (Low): Task 7 mutation 11 names the wrong detector

**Location.** `:2630-2632`.

**Defect.** Setting `Vary: Cookie` on every response puts it on both the
gate's 404 and the terminator's 404, so the members indistinguishability
row still passes. The G7 row "no 404, JSON answer, asset, or stylesheet
carries it" (`:2573`) is the row that kills this mutant.

**Proposed fold.** Name the G7 row as the detector.

### A7 (Low): Task 5 mutation 7 cannot exist under the variant design

**Location.** `:2371-2373`.

**Defect.** The variants are fixed at startup per stampable palette, so
no `dubplate-gruvbox` body exists to stamp. The mutant is only
expressible as a per-request rewrite, which the byte-identical-variant
row already kills.

**Proposed fold.** Restate it as "build the variant set from the
`[a-z0-9-]+` rule instead of the extraction", or withdraw it the way V7
withdrew mutation 2.

### A8 (Low): one cookie state row is missing

**Location.** Task 5 cookie rows (`:2311-2322`).

**Defect.** The two-cookie row proves only first-parsed-wins among
parseable values. `readCookies` also *skips* an unparseable same-name
cookie, and R-5cii22 states this, but no row pins it: a first
`__Host-dubplate_theme=a\b` followed by `label-night` stamps
`dubplate-label-night`. A padded value (`= label`, where the space
survives inside the value) is also unstated.

**Proposed fold.** Add the skip row, and optionally the padded row, which
resolves to the default. Every other cookie state the brief lists is
present, and each named mutation (8, 9, 10, 11) is killed by a named row.

### A9 (Low): Task 5's pre-named split pushes Task 6 to five deliverables

**Location.** `:2403-2406`.

**Defect.** The split moves the cookie reader to the head of Task 6,
which already carries four deliverables at `opus`. That breaks the
four-deliverable reasoning the amendment itself uses to defer the
`theme-color` metas (`:1433-1436`).

**Proposed fold.** Make the split target its own task, 5b, between Tasks
5 and 6.

### A10 (Low): the exit gate omits `Vary: Cookie`

**Location.** Exit item 8 (`:2780`).

**Defect.** Task 7 proves `Vary: Cookie`, but no exit item names it.

**Proposed fold.** Add it to item 8, citing Task 7.

### A11 (Low, for Geoff at leisure): W5.9 "takes no erratum" is an interpretation

**Location.** Spec deviations, R-5cii22 (`:273-279`), and `:1293-1298`.

**Defect.** W5.9's text reads "A system-following pick is a third value
the Go layer resolves and stamps". Here the Go layer stamps and CSS
resolves. W5.9's two reasons do not apply to this block: it is not dead
code, and it sits at (0,1,0) in a layer. The plan's reading is
defensible, but it reinterprets a ratified line.

**Proposed fold.** List a one-line clarifying erratum in Task 8's errata
rather than "no erratum".

## Task order and cross-section consistency

- **Build order across Tasks 4 to 6.** Task 4 supplies `system.css`
  before Task 5's extraction asserts `{label, label-night, system}`.
  Task 5 adds the meta to `app.html`, which Task 3 created. Task 6's tier
  links follow the variants. Each commit can be green once A1's row is
  reworded.
- **Exit items 7 and 15.** Both agree with Tasks 4 and 5.
- **The deferred list.** It matches R-5cii22's "what lands where".
- **The goal.** It agrees with the privileged-fields theme row, which now
  names both channels and the cookie's refusals.

## For Geoff

Nothing blocks. The amendment decided these without Geoff:

- the adopter-facing default `DUBPLATE_THEME_PALETTE=system`, which the
  record flags
- `--size-field` at `0.275rem` against his draft's `.25rem`, which the
  record flags
- branding flattening both palettes (A4, unflagged)
- moving the `theme-color` metas to 5c-iii (recorded, and reversible)

Also for Geoff: the W5.9 reading (A11).
