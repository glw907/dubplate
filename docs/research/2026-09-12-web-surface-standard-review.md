# Web surface standard, revision 1: adversarial review and fold record

The document: `docs/specs/2026-09-12-web-surface-standard.md`.
Seven lenses read revision 1 against the brief in the session scratchpad:
A citations, B contradictions with dubplate's ratified documents, C stack
correctness with web research, D upgrade resilience, E gate integrity and
proportionality, F accessibility standards, G product fit and scope.

This file lists every finding by id with its disposition. Revision 2 of
the standard is the fold. The harvest evidence, the per-clause citations,
the reversals list, and the declined table moved to
`docs/research/2026-09-12-cairn-harvest-reference.md`.

## Totals

| Severity | Findings | Folded | Folded with adaptation | Rejected |
|---|---|---|---|---|
| Blocker | 22 | 22 | 0 | 0 |
| Major | 59 | 52 | 7 | 0 |
| Minor | 55 | 50 | 5 | 0 |
| **Total** | **136** | **124** | **12** | **0** |

Nothing was rejected. One finding (E-18) recorded two claims as verified
sound and asked for no change; it is counted as folded, meaning its
conclusion was carried rather than re-checked.

By lens: A 16, B 30, C 17, D 16, E 18, F 21, G 18.

## Clause numbering

**No W clause number is retired.** Every revision 1 clause survives in
some form. Eight clauses are new: W5.22, W5.23, W6.10, W6.11, W7.18,
W7.19, W12.9, W13.9.

**Sections 17 (reversals not imported) and 18 (what was declined) are
retired as section numbers.** Their content moved whole to the reference,
and section 19 (supersessions and errata) is new rather than a reuse of
either number.

Four gate rules named in revision 1 are dropped rather than renumbered,
because they are rules and not clauses: `norms-bands`, `panel-width`,
`screen-anatomy` (replaced by a small `h1-count` rule of dubplate's own),
and `gap-scale` (replaced by `role-utility`). `token-resolution` is
dropped as a duplicate of DL G2. `relational-spacing` is deferred past
rung 7 rather than dropped.

## Two conflicts between lenses, and how each resolved

**F-6 against C-1 and D-1, on where the guarded-control override lives.**
F-6 proposed pinning a `@layer components` rule to restore
`pointer-events`. C-1 and D-1 both compiled the stack and showed that a
components-layer rule cannot outrank a DaisyUI declaration, because
DaisyUI emits into sublayers of `utilities`. The compiled evidence wins:
W7.8's override is unlayered and sits in W5.22's quarantine file. F-6's
substance (the override is required at all, the handler must no-op, the
reason must be keyboard-reachable) folds unchanged.

**B's m-6 against E-6, on the `token-resolution` rule.** m-6 proposed
rewording it to cover `@theme` blocks. E-6 read the engine and found no
such rule exists, and that DL G2 already asserts each theme file's exact
declared name set. E-6's evidence is the stronger reading, so the rule is
dropped rather than reworded.

## Lens A: citation fidelity (5 major, 11 minor)

| Id | Severity | Disposition |
|---|---|---|
| CIT-1 | major | Folded. W7.6 no longer counts hand-rolled copies; the reference cites CI:155-156 without a number. |
| CIT-2 | major | Folded. W12.4 carries a tier column and marks `chip-ground-collision`, `field-edge-alignment`, `container-inset-asymmetry`, `form-font-parity`, and `weight-budget` advisory, with cairn's 60% false-positive ruling named. |
| CIT-3 | major | Folded. W4.12 carries the padded-chip carve-out in cairn's own words: the chip is a box and its padding-box center is the reading. |
| CIT-4 | major | Folded. W3.7's "never mounts" is scoped to an unmatched path, with the thrown in-route `error()` named as the other code path. |
| CIT-5 | major | Folded. The reference splits the citation: GRADE for the capture set and reader's discipline, the craft-chapter acceptance record for k=3 and the 2-of-3 bar. |
| CIT-6 | minor | Folded. `exemplar-list.md:92-94`. |
| CIT-7 | minor | Folded. ADS:471 replaces ADS:466-467 in W3.2. |
| CIT-8 | minor | Folded. The reference's `TK` row covers `src/lib/design/` and `scripts/build/`, and names the four misaddressed files. |
| CIT-9 | minor | Folded. The 51/120 pair cites the vertical-alignment harvest findings. |
| CIT-10 | minor | Folded. W7.9 scopes the hairline: universal for the tint families, dark-only for `.btn-active`, where 1.14:1 was measured, with the light-scheme measurements stated. |
| CIT-11 | minor | **Folded with adaptation.** Rather than pick one cairn passage, W4.4 makes the stacked gap rung 7's choice and records that cairn's two passages disagree. This is lens A's second proposed fix. |
| CIT-12 | minor | Folded. ADS:43-44 added. |
| CIT-13 | minor | Folded. ADS:518 added for the two-pane collapse. |
| CIT-14 | minor | Folded, both halves. W5.17 scopes the 5.6 change to `.btn`, and its closing paragraph plus W12.3's origin column say which rows the ported rule covers and which dubplate adds. |
| CIT-15 | minor | Folded. PDS:153-157. |
| CIT-16 | minor | **Folded with adaptation.** W2.8 keeps the styleguide as the third reader and states plainly that the substitution is dubplate's, since dubplate has no reference docs page. |

## Lens B: contradiction with dubplate's ratified documents (7 blockers, 11 majors, 12 minors)

| Id | Severity | Disposition |
|---|---|---|
| B-1 | blocker | Folded. W7.15 states fork 4's kind answer, the admin contact, the threshold flag, and the three authored outcome strings. |
| B-2 | blocker | Folded. W3.7 is rewritten to the client-router shape: the Go mux's 404 terminator stands as a trust boundary, and the app's own 404 is a client-side navigation state. |
| B-3 | blocker | Folded. `themes: false` is removed; W5.9 states why DL G10 and R3.4 need the built-ins emitted. |
| B-4 | blocker | Folded. The `:root:not([data-theme])` guard and the `prefers-color-scheme` fallback block are removed from W5.9 and W5.10; a system-following pick is a third stamped value. |
| B-5 | blocker | Folded. W5.10 says rung 5 builds no picker and the clause lands at rung 7; section 15 routes it there. |
| B-6 | blocker | Folded with E-8 and G-9. W12.6 moves the rendered leg out of the per-task gate and names the port as an environment variable and the browser as a hard requirement. |
| B-7 | blocker | Folded. W2.7 is additive: R2.6's third condition and its exception line stand unchanged, and the ink utility covers dubplate's own markup only. |
| M-1 | major | Folded. The operator-overlay clause is dropped from the dual-gamut rule; a serve-time `doctor` advisory is named as the honest form. |
| M-2 | major | Folded. W2.6 scopes to the generated branding stylesheet through DL G8, and states the operator file as outside every gate's reach with W5.11 as the only defense. |
| M-3 | major | Folded with G-5 and E-13. Section 15 cuts 5c to its ratified deliverable list and states the 4.5M ceiling is unchanged. |
| M-4 | major | Folded. The ink tokens are declared outside every theme block as grammar-layer tokens under the palette's name, so DL G2's exactly-21 arithmetic is untouched. |
| M-5 | major | Folded. W11.5 is rung 7's uploader coherence gate; the standing `see-it-before-geoff-tests-it` bar stays the lane's own render and read. |
| M-6 | major | Folded. W1.4, W8.9, and W12.3's `quote-grounding` row all name `research/passb-copy-final.md` as the authority, and the rule now checks shipped strings against the deck rather than the reverse. |
| M-7 | major | Folded. W4.10 drops the cairn `btn-sm`/`btn-xs` justification, keeps 44 as the requirement's word (SC 2.5.5), and rests the 24px floor on the unmeasurable-by-token argument. |
| M-8 | major | Folded. W11.8 is a four-layer fence whose fourth layer is a Go build tag excluding the session-minting handler from the release binary. |
| M-9 | major | Folded. W8.8 requires the client map to carry failure kinds and render Go-supplied wording, or to be generated with a no-drift test. |
| M-10 | major | Folded. Section 19 records seven supersessions, each owed as an erratum in the document named. |
| M-11 | major | Folded. W1.1 and the header claim no palette value, face, or shipped product string, and admit measured floors and labelled exemplars. |
| m-1 | minor | Folded. `dubplate-server serve` throughout. |
| m-2 | minor | Folded. "Sections 2 through 13". |
| m-3 | minor | Folded. Section 14 is the single list and W10.8 points at it. |
| m-4 | minor | Folded. W4.10's conductor ruling cites W12.4's row and W12.2's tiering. |
| m-5 | minor | Folded. The `achromatic` row is rescoped to the palette files. |
| m-6 | minor | **Folded with adaptation.** The rule is dropped, not reworded; see the conflict note above. |
| m-7 | minor | Folded with M-6. W1.4 re-sourced. |
| m-8 | minor | Folded. Section 15's hook sentence is scoped to this standard, with DL's hook 6 named as the counterexample. |
| m-9 | minor | Folded. Section 15 lands the AI-tells rule at 5c, where R5 already requires a register gate, and every other W12 rule at rung 7. |
| m-10 | minor | Folded. W11.3 is admin-only, and the route lands at rung 7 beside W11.8's fence. |
| m-11 | minor | Folded. W4.1 binds the both-schemes sweep from rung 7. |
| m-12 | minor | Folded. W4.7 is retitled "dialogs on a narrow screen" and cross-references W5.14. |

## Lens C: stack correctness (3 blockers, 6 majors, 8 minors)

| Id | Severity | Disposition |
|---|---|---|
| C-1 | blocker | Folded with D-1. W5.18 is one pinned unlayered rule with a version stamp, quarantined in W5.22's file; W5.8 states the measured layer order. |
| C-2 | blocker | Folded. `data-theme` lives on the document element per DL R4.2, and W5.9 records why cairn's wrapper rule does not transfer. |
| C-3 | blocker | **Folded with adaptation.** Neither of the two named routes was taken: revision 2 adopts lens C's own recommendation from its open-question section, dropping the inline script for a cookie the Go server reads while rewriting the document. W5.10 states the signed-out and pre-rung-6 path plainly. |
| C-4 | major | Folded. W9.3 binds authored CSS, and the reduced-motion guarantee is a rendered assertion over a global kill switch, itself an intentional unlayered entry. |
| C-5 | major | Folded with F-3. Preflight is kept, and W5.16 becomes an affirmative `role="list"` rule with the `.list` row rewritten for the Preflight-present world. |
| C-6 | major | Folded with F-6. W7.8 ships the unlayered `pointer-events` restoration and the DaisyUI note moves into W5.17's correct row, scoped to `[aria-disabled="true"]` on `.btn`. |
| C-7 | major | Folded. W2.5 is narrowed to what `@theme` actually emits, with `@utility` required outside a recognized namespace and the definition test asserting each role's utility. |
| C-8 | major | Folded. W5.13 refuses the CSS-focus open method, not the class, and the gate rule keys on the mechanism. |
| C-9 | major | Folded. W3.7 states that the build emits one prerendered document per owned route and that DL R4.5's rewrite and R5.6's hash set cover each. |
| C-10 | minor | Folded. W5.21 states that dubplate keeps Tailwind's automatic source detection and that the clause reduces to the uncompiled-class gate. |
| C-11 | minor | Folded. W7.17's inert bar is a plain `<div>`, and `role="presentation"` on `<progress>` is named as non-conforming. |
| C-12 | minor | Folded. `badge-soft`, `badge-dash`, and `.validator-hint` join W5.17's table; W6.6 routes the hint through W2.7's ink. |
| C-13 | minor | Folded with F-19. W4.8 bounds the `contents` wrapper to a semantics-free `<div>` and adds a rendered accessibility-tree assertion. |
| C-14 | minor | Folded. W5.4 requires `{@attach}` for DOM-lifecycle behavior, with `fromAction` for action-only libraries. |
| C-15 | minor | **Folded with adaptation.** Rather than a new clause, section 14 names forced-colors as explicitly undecided and points at Playwright's emulation as the cheap seventh capture. Lens C offered this as its alternative. |
| C-16 | minor | Folded. W2.10's third bullet says "that element's own `mt-*`". |
| C-17 | minor | Folded. The reference's reversal entry adds the Baseline support line as the durable reason. |

## Lens D: upgrade resilience (2 blockers, 10 majors, 4 minors)

| Id | Severity | Disposition |
|---|---|---|
| D-1 | blocker | Folded with C-1. W5.18 unlayered, named in W5.8's pinned set and W13.2's version-coupled list. |
| D-2 | blocker | Folded. W13.1 states that neither variable names nor class names are stable, and W13.9 states that a DaisyUI major is a `web/` plus `internal/config` plus adopter-branding event with a named migration obligation. |
| D-3 | major | Folded. W5.22 quarantines every DaisyUI-internal-coupled override in one file with version stamps, and W5.7 forbids scoped styles from reaching a DaisyUI class. |
| D-4 | major | Folded. W13.3 diffs declarations, not only presence, with the `outline-none` case as the evidence and a rendered assertion on the accessibility-meaning properties. |
| D-5 | major | Folded. W13.2 gives the version-coupled set one committed home and a gate that fails when a stamp does not equal the pin, and extends the list to eleven entry groups. |
| D-6 | major | Folded. W2.10's first bullet is the general rule across Tailwind's `--container-*` key set, with a static rule. |
| D-7 | major | Folded. Every W5.17 row carries the version it was measured against and a falsifiable test that renders the stock default. |
| D-8 | major | Folded. W13.3 narrows the inventory to the classes dubplate's markup writes, with cairn's public-API reason named as what dubplate does not inherit. |
| D-9 | major | Folded. W13.5 makes the styleguide the named home of the committed baseline set plus a geometry assertion set. |
| D-10 | major | Folded. W13.9 states plainly that DaisyUI is load-bearing and that a different component library is a redesign. |
| D-11 | major | **Folded with adaptation.** No full wrapper set. W5.23 mandates one named recipe per coupled control family, states the expected count of six to nine, and cites the retired cairn wrappers as the evidence against going further. This is lens D's own recommended narrow version. |
| D-12 | major | Folded. W13.6 replaces the blanket claim with the four practices' actual reach and names the SvelteKit 3 RC surfaces, the adapter Vite plugin seam included. |
| D-13 | minor | **Folded with adaptation.** The scheduled watcher is replaced by lens D's option (a): a `renovate` or `dependabot` config in this repo restricted to major bumps of the four UI packages, routed to rung 7 as one committed file. |
| D-14 | minor | Folded. W13.4 splits the six steps into assertion gates and a baseline review, and drops "one commit" for one branch. |
| D-15 | minor | Folded. W13.7 gains the `pin-exactness` gate, and the "nowhere else" wording is reconciled with W13.2's stamps. |
| D-16 | minor | Folded. W13.8 requires the release note and the binary's upgrade output to name the operator CSS file after a major. |

## Lens E: gate integrity and proportionality (4 blockers, 10 majors, 4 minors)

| Id | Severity | Disposition |
|---|---|---|
| E-1 | blocker | Folded. W12.8 flips: a port is the primary path and the dependency is not evaluated for the rendered rules at all, with the identity-comparison mechanism stated. |
| E-2 | blocker | Folded. W12.1 adds the un-booted page shape and the per-page expected-identity assertion at exit 2. |
| E-3 | blocker | Folded. W12.1 adds the stale-embed shape, and W12.6 requires the leg to build, rebuild, and compare bundle hashes. |
| E-4 | blocker | Folded. `gap-scale` is dropped for `role-utility`, which gates W2.4 directly; `type-scale` and `grammar-boundary` are named in the reference as port-only with the token prefix as the fork point. |
| E-5 | major | Folded. W12.8 states the true count (ten rules and three harness constants) and the reference enumerates them, `norms-bands` included as a rule that would import cairn's look. |
| E-6 | major | Folded. W12.3 carries an origin column separating engine-provided rules from dubplate's own builds, and notes cairn's `scripts/` is not in the published package. |
| E-7 | major | Folded. W12.4 names all four unused rendered rules and why each is unused. |
| E-8 | major | Folded with B-6 and G-9. W12.6 routes the rendered leg to the pass-end gate, CI, and paint-moving or `web/`-touching tasks, citing the 2026-09-09 gate-economy ruling as applied rather than excepted. |
| E-9 | major | Folded. W12.6 names the browser as a hard-required tool failing by name, the way DL R13 treats Node. |
| E-10 | major | Folded. W9.2 and W9.3 are restated as reviewer-enforced with a named narrow static half, and the reduced-motion guarantee moves to a rendered assertion. |
| E-11 | major | Folded with CIT-14. W5.17 and W12.3 state which rows the ported rule covers and that the guarded-control marker class is renamed. |
| E-12 | major | Folded. W12.1 states that the login path is configuration and that an unauthenticated rendered run can audit only the auth pages until W11.8's fixture exists. |
| E-13 | major | Folded with M-3 and G-5. Section 15 cuts 5c to its ratified deliverables and states the ceiling is unchanged. |
| E-14 | major | Folded. Every clause outside section 12 carries a `G`, `R`, or `g` mark, and the header states that an unmarked clause is a defect the next revision closes. |
| E-15 | minor | Folded. W12.3 says `.svelte` files under `web/src` and names the TypeScript limit. |
| E-16 | minor | Folded. W12.2 requires dubplate's own suppression marker and states the switching cost. |
| E-17 | minor | Folded. W12.8 states the dependency's true surface and defers the adoption-bar evidence to rung 7's plan rather than asserting it. |
| E-18 | minor | Carried, no change needed. The two W12.1 claims lens E verified (exit 2 on a run that cannot start, a nonexistent scope path failing) stand as written. |

## Lens F: accessibility standards (2 blockers, 9 majors, 10 minors)

| Id | Severity | Disposition |
|---|---|---|
| F-1 | blocker | Folded. W4.10 carries SC 2.5.8's five exceptions by name, implements Inline and Spacing before anything fails, restates the measurement as a 24 by 24 square fitting wholly inside the activation region, and sweeps every W4.1 viewport. |
| F-2 | blocker | Folded. W5.15 puts `aria-sort` on `<th scope="col">` with the header text wrapped in a button, plus a static rule asserting it never appears on a non-header element. |
| F-3 | major | Folded with C-5. W5.16 is affirmative, `role="listitem"` is scoped to the display-change case, and the `<nav>` exception is recorded. |
| F-4 | major | Folded. W4.5 corrects the "every engine" claim with the Chrome 80 / Firefox 62 / Safari 17 history and rests the ruling on `display: contents` and the folding recipe. |
| F-5 | major | Folded. W7.17 makes a real `progressbar` with an accessible name the default, voiced through a throttled `role="status"` region for SC 4.1.3, with the inert bar narrowed to an ornamental case. |
| F-6 | major | Folded with adaptation of the mechanism only; see the conflict note. W7.8 pins an unlayered override, requires the handler to no-op, and routes the reason through W7.6's live region so it is keyboard-reachable. |
| F-7 | major | Folded. W7.6 specifies clear-then-set as the mechanism, names U+00A0 if a nonce is kept, and marks the measurement as owed rather than made. |
| F-8 | major | Folded. W6.2 puts hints and error lines outside the wrapping label with `aria-describedby`, and names the `aria-labelledby` escape when layout forces one inside. |
| F-9 | major | Folded. W7.18 names SC 3.3.8 and technique G218, forbids a transcribed-code variant without paste, and forbids paste suppression on any credential field. |
| F-10 | major | Folded. W6.10 requires a single-pointer equivalent for every drag affordance, with a rendered assertion that a page carrying a drop target carries a file input. |
| F-11 | major | **Folded with adaptation.** W12.9 maps every criterion this standard discharges; W6.11 lands 3.3.7 and 1.3.5; W12.4 adds a `text-spacing` advisory rule for 1.4.12; 2.2.1's security exception is recorded in W12.9. SC 2.4.7's "visible enough" half is left to W11.5's coherence read rather than given a new rendered threshold, because no threshold was measured. |
| F-12 | minor | Folded. W3.6 states that no browser moves focus to a non-focusable fragment target and names WebKit bug 17450, and adds SC 2.4.1. |
| F-13 | minor | Folded. W7.7 requires `aria-describedby`, forbids `aria-modal` on a native `<dialog>`, and bounds the Escape suppression to the in-flight operation with a visible busy state. |
| F-14 | minor | Folded. W5.13 states the user agent supplies `aria-expanded` on a `popovertarget` invoker, names `aria-haspopup` for a menu panel, and demotes `aria-controls`. |
| F-15 | minor | Folded. W5.15 states the no-selection tab stop and marks Home and End as dubplate's addition. |
| F-16 | minor | Folded. W7.10 states the `title` is a pointer-only enhancement and that no information exists only there. |
| F-17 | minor | Folded. W7.14 requires the focus move after `tick()` alongside the attribute, and requires each state to pick moving focus or announcing. |
| F-18 | minor | Folded. W7.9 names both measured pairs and routes them into `border-contrast` at error tier for that case. |
| F-19 | minor | Folded with C-13. |
| F-20 | minor | Folded. W4.5 specifies `tabindex="0"` applied only while the content overflows. |
| F-21 | minor | Folded. W2.7 names SC 1.4.3 and its large-text allowance and states that dubplate holds every status ink to 4.5:1 regardless. |

## Lens G: product fit, personality, and scope (4 blockers, 8 majors, 6 minors)

| Id | Severity | Disposition |
|---|---|---|
| G-1 | blocker | Folded. W6.4 exempts a single-action form and moves the row's alignment to rung 7. |
| G-2 | blocker | Folded. W7.13 binds 2.0's web admin surface; section 15 states that rung 14 lands the API and takes no clause from this standard. |
| G-3 | blocker | Folded. W3.4 keeps the pinned topbar height and its reason and cuts the fill, hairline, and no-shadow ruling, naming `--depth` as the theme's. |
| G-4 | blocker | Folded. W7.11 is retargeted at rung 13's admin decision queue, and states plainly that the uploader has no accept sweep. |
| G-5 | major | Folded with M-3 and E-13. |
| G-6 | major | Folded. W3.3 binds list and settings screens; single-purpose form screens and direct-manipulation screens declare their own anatomy, and only the `h1` count binds every page. |
| G-7 | major | **Folded with adaptation.** Rather than standardizing the uploader's interactions here, W7.19 names the four structures (drop-target states, the file-picker route, multi-file progress, the self-hiding collection picker) as rung 7's, which is lens G's second proposed fix, and section 14 lists them among what this standard does not decide. |
| G-8 | major | Folded. W7.10 states the register set's size and membership are rung 7's, derived from the eight-state journey vocabulary, and that cairn's three are not inherited. |
| G-9 | major | Folded with B-6 and E-8. |
| G-10 | major | **Folded with adaptation.** Per the conductor's fold rule 2, the home is reopened rather than deleted: W1.3 keeps theme names as a personality home on the one axis DL R3.6 leaves open, the `dubplate` match block's own name and any palette dubplate authors rather than adopts. |
| G-11 | major | Folded. W1.3 names three surviving homes with one worked illustration each, labelled as examples rung 7 may discard, and adds the rule that a personality choice outranks a harvested mechanic, with section 19 as the amendment's home. |
| G-12 | major | Folded. The harvest evidence, per-clause citations, reversals, declined table, and the longer arguments moved to the reference; section 15 names the clauses that bind each rung. |
| G-13 | minor | Folded. W4.7's cap values and W8.6's separator glyph are rung 7's, and W7.2 reads "a mark or illustration, concept-named copy, and the surface's own primary action". |
| G-14 | minor | Folded. W8.2 carries dubplate's own leak list, with MusicBrainz and collection named as the two context-dependent cases, and softens the personification ban to the honesty rule under it. |
| G-15 | minor | Folded with D-13. W13.4 names the mechanism, states it is workstation and CI only and never a box process, and routes it to rung 7. |
| G-16 | minor | **Folded with adaptation.** The reference states the cairn commit (`446d5d5`, package 0.96.0) as the anchor and gives the re-verification procedure once. A public repository URL is named as the resolution path rather than substituted for the local path, because the repository is not public today; the commit is what makes a citation checkable later. |
| G-17 | minor | Folded. The row is renamed `ai-tells` and states that it enforces no register, with the release-time `--list` read named. |
| G-18 | minor | Folded. W11.8 states the assertion: the embedded bundle is searched for the fixture's sentinel and the leg fails on a hit, and a release build fails if the tagged handler is reachable. |

## What the fold could not achieve

**Length.** The conductor asked for revision 2 to be shorter than
revision 1 once the reference took the harvest prose. It is not: 1719
lines against 1293. The reference absorbed about 420 lines of evidence
and the retired sections 17 and 18, but the folds added more than that
back. The growth is concentrated where the lenses demanded new normative
content: section 12 grew by 91 lines (the origin and tier columns, the
three new silent-green shapes, the leg-placement ruling, the conformance
map), section 13 by 66 (the version-coupled home, the declaration diff,
the not-swappable clause), section 5 by 75 (the override file and the
control recipes), and section 7 by 67 (the three new accessibility
clauses). Eight new clauses and section 19 account for about 120 lines of
that. Cutting to 1293 would mean dropping folded blockers or majors,
which fold rule 1 forbids, so the length is reported rather than paid
for. If Geoff wants the standard shorter, the cheapest honest cut is
moving section 12's two rule tables into the reference and leaving the
doctrine clauses in the standard, which would save about 40 lines and
cost a rung implementer one file lookup.

**Two measurements are owed rather than made.** W7.6's live-region
re-announcement mechanism needs a standing proof test naming the
screen-reader and browser pairs; revision 2 says so rather than asserting
a character's behavior. W4.10's 44px error tier waits on Geoff, per the
conductor ruling.

## Revision 2.1 fold

The prose and fact review of revision 2 (register: Google developer
documentation, web dialect; deterministic floor recorded in the review)
returned 29 findings: 6 blockers, 10 majors, 13 minors. All 29 are
folded. Conductor rulings that changed a proposed rewrite are noted on
the row.

### Blockers

| Id | Disposition |
|---|---|
| BL-1 | Folded, with a ruling. The cookie stays; it is a deliberate supersession of DL R4.2's `localStorage` ruling, because the `localStorage` pick needs an inline script the CSP blocks. W5.10 now says it supersedes R4.2's store for the signed-out path and gives the reason, and section 19 carries the R4.2 row. Section 16 gains a fifth item so Geoff reads the supersession. |
| BL-2 | Folded, reshaped by ruling. W13.2 is no longer an equality gate across entries. Each stamped value names the version it was measured against, entries legitimately differ, and the bump rehearsal re-measures every entry behind the pin. |
| BL-3 | Folded, reshaped by ruling. Section 15 now splits new artifacts from properties: only new artifacts (grammar module, live-region helper, styleguide, dev fixture, the rendered rules) move to rung 7, and the property clauses in sections 4, 5, 7, 10 and W4.10 bind 5c's three pages as built. The 5c ceiling is unchanged and the section says why a property is not a deliverable. |
| BL-4 | Folded, reference corrected. W12.6 now cites DL R1.5 and DL section 6's preamble, with Node a hard requirement at 5c per R5's 5c row. |
| BL-5 | Folded, reshaped by ruling. Section 15 states that the AI-tells rule is additional and enforces no register, and that R5's Microsoft register gate inside 5c stands as R5 states it. |
| BL-6 | Folded. The "stands as" line left with MI-4's dropped section 19 row; the document now carries no instance. |

### Majors

| Id | Disposition |
|---|---|
| MA-1 | Folded, direction corrected. W5.14 now says the inner `<form>` start tag is ignored and the dialog's controls join the page form. |
| MA-2 | Folded, reshaped by ruling. W2.7's inks are named as a third layer, the measured status inks, in a per-palette ink block selected by the composed `data-theme` value. They are not grammar tokens and not palette-block names. |
| MA-3 | Folded. W2.7 and W5.17's row now claim only `--color-warning` as measured (about 2.2:1, cairn ADS:207-215); success and error are measured at rung 7. W13.2's entry follows. |
| MA-4 | Folded, versions removed. W5.13 says anchor positioning is not yet Baseline across current evergreen browsers and is checked at rung 7, with the `WATCH:` comment naming that table. |
| MA-5 | Folded, reshaped by ruling. The display-p3 leg is dropped to a rung 7 evaluation note rather than a gate, since DL R2.6 fixes sRGB. The rule is renamed `contrast-extension` in W12.3 and in W2.7's mark. |
| MA-6 | Folded. W12.6 states the gate at about two minutes and a bare Playwright floor at 9.2 seconds (12.8 with node start), and says the real leg's total is unmeasured until rung 7 builds it. |
| MA-7 | Folded. Sections 6, 7, and 8 gain preamble defaults: `daisyui-a11y-reviewer` for 6 and 7, the prose gate plus a human copy read for 8. The sixteen bare `R` marks now resolve. |
| MA-8 | Folded. W5.10 becomes `R` web-auth-security-reviewer plus a named gated half (no `<script>` beyond SvelteKit's generated set) and says why G6 does not gate it. W5.9 becomes `G` DL G7 for the precedence half and `R` svelte-reviewer for the rest. |
| MA-9 | Folded. W3.7 states DL R4.5's mechanism in the singular and says the same rewrite covers each prerendered document; section 19 carries the R4.5 widening row. |
| MA-10 | Folded. W13.6 attributes the SvelteKit 3 specifics to the Svelte team's September 2026 release notes, marks them RC-stage, and requires a `WATCH:` comment and a migration-guide re-read before pinning. |

### Minors

| Id | Disposition |
|---|---|
| MI-1 | Folded. "R36" is now "R5's R36". |
| MI-2 | Folded. W12.3 names cairn's four checks and puts them under `scripts/`. |
| MI-3 | Folded. "for about a year" is now "across the migration and several releases after it". |
| MI-4 | Folded. Section 19's DL R2.6 row is dropped and the unchanged note moved into W2.7's body. |
| MI-5 | Folded. The header now reads "from the clause itself or from its section's stated default". |
| MI-6 | Folded. W12.9 cites SC 2.2.1's essential exception. |
| MI-7 | Folded. Section 12's preamble states that a `G` mark may name a committed test outside the audit engine and enumerates the eight. |
| MI-8 | Folded. The `achromatic` row states the assertion first, then the scope. |
| MI-9 | Folded. W5.18 reports 1.49:1 light and 1.77:1 dark. |
| MI-10 | Folded. W7.15 marks `noindex, nofollow` as this standard's own addition, outside the Src line's vouching. |
| MI-11 | Folded as directed. The verbless setup-colon sentences the review named are cut (W1.3, W3.7), with four more in W5.8, W9.1, W11.5, and W11.8. The legitimate enumerations are left. |
| MI-12 | Folded. W2.2's and W12.2's stacked contrast frames are each varied into a plain statement. |
| MI-13 | Folded. W4.5's "the reason is not X, it is Y" sentence is split in two. |

### Counts and effect

29 findings folded: 6 blockers, 10 majors, 13 minors. Nothing is refused
and nothing is deferred. The document goes from 1720 to 1852 lines; the
growth is section 15's rewrite (about 30 lines), W2.7's third layer and
its R2.6 note, W12.6's cost paragraph, W13.2's re-measurement rule, and
the three section preambles. Em dashes remain at zero.
