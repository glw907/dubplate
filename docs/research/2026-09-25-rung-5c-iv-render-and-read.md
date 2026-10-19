# Rung 5c-iv render-and-read (Task 10)

The first read ran on 2026-09-26 over the captures at `d3fc403`: 14 states by 4 native cells
(390 and 1440, `label` and `label-night`) plus 14 contact sheets holding the 320, 768, and 2560
cells. One fresh-context Opus reader did the read, with no plan in context, against the visual
identity spec (rev 1.3) and the web surface standard's sections 3 and 4. The error-not-found
state is excluded because it showed Go's bare 404, which Task 10b replaces. The reader read
every cell and returned the verdict **fold**.

## Defects and the conductor's ruling on each

Each defect below names its state, width, and scheme, and gives the reader's finding and the
conductor's ruling.

### Broken or misleading states

1. **members-added, all cells.** After a successful add, the reset form leaves the Streamer
   username field invalid: red, with its rule text in `--ink-error`, directly above "Added.".
   **Fold (10c):** the reset clears the invalid and `:user-invalid` state.
2. **members-duplicate-address, all cells.** The alert names the conflict, but the Email input
   shows no invalid state while the username field shows the green valid border. **Fold (10c):**
   the server error lands on the email field, through its invalid state and its
   `aria-describedby`.
3. **login-sent, all cells.** The stale lede "Sign in with your email…" still shows. The
   confirmation and "Use a different address" are centered while the column above them is left
   aligned, and the confirmation butts against the lede. **Fold (10c):** hide the lede in the
   sent moment, align the confirmation to the column, and apply W3.9 spacing.
4. **Field labels, `label` scheme.** The labels ("Email address", "Name", "Streamer username")
   measure #746e61 on #f4ead3, which is 4.22:1. That is below 4.5:1 at 11-12px. On
   `label-night` the same labels measure 5.72:1, which passes. **Fold (10c):** raise the label
   ink until it passes 4.5:1 on every base of both themes, and add a gate row that measures it.
   VI rev 1.3 already raised placeholders for the same reason.
5. **Capture determinism (harness).**
   - The primary button's fill varies between cells, from #a43f39 to #c8675f against the resting
     #c23f36, so a transition or a pending state is being captured.
   - The member count leaks from cell to cell (4 → 21), so one state no longer appears at ten
     sizes.

   **Fold (10b):** settle transitions and pending states before each capture, reset the roster
   fixture per cell (or give each cell its own fixture), and assert the member count in the
   state check.

### Conformance to the identity spec and the surface standard

6. **W3.2 at 768, 1440, and 2560, every state.** No content measure is capped or centered.
   Rules, alerts, and lists stretch to 1376px and to about 2500px. Forms hug the left edge, and
   the refused paragraph runs about 150 characters. The signed-in header does not share the main
   column's measure: its mark sits at x=8 against content at x=32. **Fold (10c):** one capped,
   centered measure shared by header and main.
7. **W4.9 at 2560.** The reader saw no root `clamp()` scaling: the wordmark measures the same
   width at 1440 and 2560. **Fold (10c) if real:** first verify the built CSS. If `html` has no
   `clamp()`, add one per W4.9. If it has one, the task reports that and changes nothing.
8. **interstitial-refused.** The heading "That link didn't work" has a straight apostrophe.
   **Fold (10c):** curly in the deck, and a deck check that rejects straight apostrophes in
   product copy if the AI-tells rule does not already.
9. **members lists.** A member with no username shows a dangling " ·". **Fold (10c):** join only
   the segments that are present.
10. **W3.9 rhythm.**
    - Alerts sit flush under the button, and the button's shadow overlaps the alert.
    - "Add a member" sits 8px under the list.
    - On the interstitials, the title, rule, lede, and button are packed 4-5px apart.
    - The "N members" count touches the rule.

    **Fold (10c).**

### Polish and intent

11. **interstitial-signed-in is a dead end.** It has no header and no link onward. **Fold (10c):**
    verify the flow. If nothing redirects, add a deck-sourced link onward.
12. **members-non-admin.** It renders "Page not found" with no header and no sign-out. **Fold
    (10c):** keep the not-found body (register-gate ruling: the tab never says Members to a
    non-admin), and render the signed-in header without a Members link so the member can still
    sign out.
13. **Phone form gap.** About 90px separate the last field from the button at 390, against about
    37px at 1440. **Fold (10c)** if a reserved hint slot causes it. Otherwise record why.
14. **A valid (green) border beside a server-failure alert.** **Refused:** the field is valid,
    and this is polish that costs more than it gains.

The reader found the following conformant: the tokens, inks, reversed label, keylines, hard
shadows, ADMIN badge, toggle, neutral header, and link ink. No 320 cell overflows.

## Routing

Task 10b (Opus) takes the harness items: the 404 fallback page, the deadlines, the
inline-anchor and favicon narrowing, and item 5. Task 10c takes the page items 1-4 and 6-13.
The matrix then re-runs, and one fresh read of the changed cells, plus error-not-found, comes
before Geoff's look.

## Task 10b's escalation and the ruling (2026-09-26)

Task 10b (`47331b4`) landed criteria 2 to 5: the deadlines, capture determinism, and the two
narrowings. Criterion 1, the styled not-found fallback, did not land. `RequireAdmin`
(`internal/auth/admin.go:112`) answers a refused request exactly as `http.NotFound` does. That
way a non-admin cannot tell a gated path from an unregistered one, and two `serve_test.go` tests
enforce it. So a styled 404 at unknown paths also needs the admin gate's refusal routed through
the same handler. That is a change to the rung 5 auth surface and needs its own security review.
The fallback document also needs its links rebased under the prefix, and its inline-script CSP
hash recomputed.

**Ruling: deferred to rung 7.** A bare 404 on an unknown URL is cosmetic, and today it is
indistinguishable by design. Rung 7 already owns the error pages' visuals (Ruling 3's
error-page photos). The unlanded work is kept at
`docs/superpowers/research/patches/2026-09-26-styled-404-fallback.patch`, and the rung 7 task
starts from it. That task adds an `AdminConfig.NotFound` (or equivalent) so the gate's refusal
and an unmatched path answer the same styled document. It runs under a
`web-auth-security-reviewer` read. For 5c-iv, the error-not-found capture asserts the bare
404 and keeps `skipUIFloors`.

## The re-read after the fold (2026-09-26, captures at `c1bec33`)

One fresh-context Opus reader read all 75 images: every native cell and every contact sheet,
error-not-found included. It found nothing broken. Every sampled text pair clears 4.5:1, both
schemes hold at all five widths, and no cell overflows. Its verdict was **fold**, with rulings:

1. **The slab ladder is missing.** Every `h1.display.page-title` inherits 16px, so a page title
   renders smaller than the 18px wordmark, and login-sent's h2 matches the h1. The ladder is
   h1 36, h2 20/24, and wordmark 24 in the header and 30 on login
   (`2026-09-24-identity-refinement/ledger-typography.md:28`, which the identity spec calls
   "already right"). `.display` has no size (`components.css:408`). **Fold (5c-iv's review
   fold)**, with a rendered row that measures the ladder.
2. **Identifiers break at their own hyphen** ("capture-" / "added"), and meta lines break just
   after " · ". **Fold:** keep each segment together, as the identity spec's catalog-line row
   already asks.
3. **The login-sent ghost button's ink sits 19px inside the column edge**, so it reads as
   indented bold text. **Fold:** pull its ink onto the column edge, or give it a link
   treatment. The header toggle's 11px inset is too small to act on.
4. **The interstitials carry no mark or group name.** **Deferred to rung 7:** the spec does not
   mandate the lockup here, and rung 7 authors identity visuals.
5. **A green valid border beside "We don't recognize that address".** **Refused:** it falls
   under item 14's ruling. The address is well-formed, so the field is valid.
6. **The field-to-button gap and the "Added." alert below the fold at 390.** **Fold, verify
   only:** confirm that focus or a live region announces "Added."; add focus if neither does.
   The gap stays.
