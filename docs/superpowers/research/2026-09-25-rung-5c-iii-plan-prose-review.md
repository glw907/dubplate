# Prose review: rung 5c-iii plan (revision 3, `28c983e`)

Reviewer: `prose-voice-reviewer` (read-only; the conductor saved and condensed its report).
Register: Google developer docs, Go arm. Checked and correct: every ceiling and flag, the 45
deliverables, the 23 outcome constants, all 16 contrast figures (recomputed with the WCAG sRGB
formula), the VI, W, and rung 5 spec pointers, and the W3C quotes (fetched). tellgrader
`--register docs`: 17,730 words, one hard-slop hit ("delve", the intentional mutation seed at
`:1442`), one bold-lead finding (the deviations list, judged a true enumeration key), cadence CV
1.01.

## Blocker

- **B1. Ruling 7's "no" offers mustard as a focus color** (`:219-220`). 7.81:1 is `label`
  secondary `#e0a526` on `#1f1b16`, which VI `:108-109` and `:113` rule "never a focus or
  selector indicator". The header-scoped `#ec7b6a` is also a literal G1 refuses outside
  `lib/themes/` (`g1.ts:11`). Rewrite: "Both samples then read the bar: 6.20:1 on `label`'s
  `#1f1b16` and 6.93:1 on `label-night`'s `#110f0d`. The header-scoped value lives in
  `web/src/lib/themes/`, since G1 refuses a literal elsewhere, and Task 9's Files gain that file.
  VI takes an erratum: the ring color is one per surface class, not one everywhere." Drop the
  mustard option.
- **B2. The `--scope-changed` premise is false** (`:706-710`, Task 3a acceptance `:1119-1120`).
  `check.sh:210-221` and `:296-313` send `--scope-changed` to `make check` whenever the change set
  touches `internal/harness/`, `internal/engine/`, or `corpus/`, so an engine-only change runs the
  rendered leg. Rewrite `:706-708`: "`--scope-changed` calls `make check` only when the change set
  touches `internal/harness/`, `internal/engine/`, or `corpus/` (`scripts/check.sh:210-221`).
  Otherwise it runs `make vet lint tidy-check` and `go test` over every package but the conform
  suite (`:296-308`), so Task 3a wires the leg into that branch explicitly." In the acceptance
  criterion, replace `internal/engine/` with `internal/atomicfile/`.
- **B3. "The last two" names the wrong inks** (`:612-614`). The inks equal to their fills are
  info `#6d9cc7` and warning `#e5ad33`; error `#e37866` differs from its fill `#e0604f`. Rewrite:
  "(the info and warning inks, `#6d9cc7` and `#e5ad33`, may be written `var(--color-info)` and
  `var(--color-warning)`, since each equals its fill; W2.7's "The ink defaults to the fill")".

## Warning

- **W1. Two surfaces ship before a caller** (`:84-88`). The `Mark` component (Task 4, `:1190`,
  `:1209-1211`) has no Task 4 row; its consumers are Tasks 8 and 9. Move it to Task 8's Files (the
  SVGs and icon links stay in Task 4), or rewrite `:87`: "The site name and the `Mark` component
  are the exceptions: their first readers are Task 8's login heading and Task 9's header." The
  `label` placeholder opacity (Task 5, `:1311`) has no consumer, since Tasks 8 and 9 forbid
  `placeholder` (`:1467`, `:1536`, `:1590`): drop it and carry "both palettes' placeholder opacity"
  to rung 7 in Task 11's handoff and the exit gate's deferred list.
- **W2. Ruling 7 reads as a measurement** (`:194-196`, `:207-209`). Question: "**7. Does the
  signed-in header's focus ring take W9.1's page-colored halo (a `base-100` halo under the
  `--ink-primary` outline, spreading past the ring), so one ring color clears 3:1 on the neutral
  bar and VI's "one focus-ring color" (`:5`) stays literally true?**" Scope the halo to
  `:focus-visible`. Replace "W9.1 supplies the mechanism" with "W9.1 prescribes a page-colored halo
  where a ground shares the ring's hue; the bar fails on luminance instead, and the same halo fixes
  it."
- **W3. Two unresolvable attributions.** `:183`: "The ceiling re-derives from the last actual, as
  5c-ii's did (5c-ii plan, "Token ceiling")." `:79`: cut "the brief's eight-task line" or name the
  brief's path.
- **W4. Task 2 mutation 2 does not isolate the pinned-value row** (`:1061-1062`). `#765000`
  measures 4.46:1 on `label` `base-300`, so G9 fails too. Rewrite: "2. Set `label`'s
  `--ink-warning` to `#744e00` (4.60:1 on `base-300`, so every contrast row still passes) → the
  pinned-value row fails alone."
- **W5. Catalog sentences** of 129 to 174 words at `:1367`, `:1100`, `:1402`, `:711`, `:10`,
  `:36-46`: turn each into a list. "What this pass is" (`:37-46`) becomes a lead sentence and a
  six-item list (faces with G4 and G5; VI 1.2 theme data with per-surface rows; the rendered leg;
  the document head; VI's component layer; 5c-iv's deck, pages, and render-and-read). Ruling 2's
  evidence sentence (`:104-110`): one fact per sentence.

## Suggestion

- Ruling 9 (`:238-240`): name the third edit, "and so rewrites `checkSumAdjustment`".
- Ruling 4 (`:138-140`): "Recommendation: yes; the quoted W3C text settles it, and VI asked for
  the ruling."
- Ruling 6 (`:186`): "six tasks at 0.725M (4.35M) plus adders".
- Pointers: `:296` `outcomeAdminNotGrantable` is `admin.go:43`; `:578` `ServeContent` is
  `document.go:230`.
- Wording: `:456-457` "not the 5c-ii plan's `5.7.45`"; `:630` "the two colors VI marks
  fill-only"; `:1724-1725` "the 2px `--ink-primary` ring at a 2px offset".
- `:966-967` against the territory wall `:443-453`: add `.gitleaks.toml` to the wall or make
  Task 1 stop and report.
- `:149`: mark W3C's "as is therefore not subject" with "[sic]".
- Ruling 8 (`:223-236`), optional: its "no" covers only "promote 44 everywhere".

Verdict: the register's plausible author. The three blockers are wrong facts, not voice; the
voice fault is the catalog sentences.
