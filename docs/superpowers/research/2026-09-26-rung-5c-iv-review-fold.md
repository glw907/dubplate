# Rung 5c-iv pass-end reviews and their fold (2026-09-26)

The three pass-end reviewers (Opus, read-only) read `392d1cd..HEAD` after the simplifier round
(`1e48f93`, which one diff-reviewer accepted). This file records their findings and the
conductor's ruling on each. The fold runs as two serial tasks: R1 covers security, the header,
and the shared plumbing. R2 covers page accessibility, typography, and the deck. The re-read's
fold items (the slab ladder, identifier wrapping, the ghost button's inset, and whether "Added."
is announced) go to R2.

## web-auth-security-reviewer (Opus): no Critical or High; two Medium
- M1 FOLD: a fragment on the mailed link (e.g. #main) makes kit's reset_focus replaceState restore ?t= to the address bar and history (reproduced in Chromium; kit 2.70.3 client.js:2026-2028, :3240-3257; page web/src/routes/login/link/+page.svelte:33-37). Fix: web/src/hooks.client.ts init hook reads t into a module holder and native-replaceState to location.pathname before start() resolves the URL (also closes L2, token in page.url); add a #main variant to interstitialOrderingRow (run.ts:2794).
- M2 FOLD: SiteHeader.svelte:66-73 signOut ignores the response and always navigates to /login; a failed clear leaves the session cookie. Fix: navigate only on response.ok and outcome signed_out; otherwise announce NETWORK_FAILURE or CROSS_ORIGIN_REFUSED in the header live region.
- L1 FOLD: run.ts:2805 ordering row snapshots the request log after load; snapshot before goto and exclude only the document GET. Optional meta referrer no-referrer in app.html: FOLD (cheap belt-and-braces).
- L2: closed by M1's hooks.client fix.
- L3 FOLD: a reload before the click shows the "works only once" refusal; give the token-absent case its own deck line ("Open the link from your email again.").
- L4 REFUSED for 5c-iv: login CSRF through a member's own link is inherent to magic links; a masked "Sign in as a***@" line needs client-side token parsing; route to the revoke-verb / rung 7 carry list.
- L5 no change: __Host- cookie is host-wide by design; value is a fixed literal; NewDocument accepts only exact Variants keys. Non-secure-context persistence failure is functional, not security.
- L6 FOLD: auth.ts:119 and members/+page.svelte:92 outcome lookups fall through the prototype chain; use Object.hasOwn.
- L7 no change: rendered test servers are build-tagged test-only loopback listeners.
- Informational to 5d's runbook: the kind unknown-address answer reveals the first admin's contact and member-vs-nonmember; operator docs must say so.

## daisyui-a11y-reviewer (Opus): one blocker, nine warnings
- B1 FOLD: Sign out on hover/focus in label: DaisyUI .btn:hover/:focus-visible paint base-200 under cream text (1.14:1 focus, ~1.3:1 hover). SiteHeader.svelte:80, components.css:218-226, :279. Fix: header ghost button uses the bar's colors (--btn-color neutral, --btn-fg neutral-content) in the component layer, or btn-neutral.
- W1 FOLD: focus drops to body when a busy button disables (login:119, link:128, members:267). Restore focus after submitting=false (tick), or aria-disabled while busy; members added -> address field.
- W2 FOLD: "Use a different address" unmounts the focused control (login:60-63); focus the email input.
- W3 FOLD: repeated identical outcome never re-announced; one shared $lib announcer (clear, set next frame) used by all three pages (W7.6).
- W4 FOLD: streamer-username format hint hidden until invalid, not described-by; persistent hint p with aria-describedby; error-only hints as validator-hint hidden (also removes the reserved gap on login and members).
- W5 FOLD: server field errors not tied to fields for blank/invalid display name, invalid/missing/duplicate streamer username, duplicate principal, login invalid_address; map each to its field (aria-invalid, aria-describedby, input-error, focus).
- W6 FOLD: login h1 empty when the site-name meta is absent; fall back to a deck title.
- W7 FOLD (same as security M2): sign-out failure silent; success unconfirmed (signed_out deck line unused).
- W8 FOLD: interstitial Continue links /members, which 404s for non-admins; show it only when the consume answer says admin, else omit and say why in the report.
- W9 FOLD: members fetch or JSON failure shows "Page not found"; separate unreachable state with NETWORK_FAILURE and retry.
- S1 FOLD: sent moment sets two sentences in Alfa Slab (spec: never body copy); slab h2 title + Work Sans p.
- S2 FOLD: outcomes announced twice (focused heading + live region).
- S3 FOLD: header nav holds no links; home link inside nav or nav -> div.
- S4 FOLD: DaisyUI-first: header on stock navbar (navbar-start/end) with a pinned height (W3.4).
- S5 FOLD: member rows fight list-row grid; list-row + list-col-grow; role=list on ul.
- S6 FOLD: header in every members state (loading too); drop the skip-link offset.
- S7 FOLD (with W5): validator on address and name fields.
- S8 FOLD: login legend sr-only.
- S9 routed to rung 7 (error page onward link, with the styled 404).
- S10 FOLD: not-found copy consistent between members refused and +error.
- S11 routed to rung 7: toggle does not follow a later OS scheme change under system.

## svelte-reviewer (Opus): no blocker; nine warnings

This review overlaps the two above on sign-out, focus loss, repeated announcements, error
classification, hints, and Continue. Those items are ruled above. Its new items:
- #5 FOLD (with W9): only a 404 means `refused`. A fetch error or any other status gets an
  error state. The post-add refresh keeps the current list and shows the failure.
- #6 FOLD (with W5): a field's error mark clears on that field's `oninput`.
- #8 FOLD: move "Skip to content", `ADD_MEMBER_LEGEND`, `ADD_MEMBER_SUBMIT`, `ADMIN_BADGE_LABEL`,
  and `REFUSED_LINK_PHRASE` into `deck.ts`. Make the refused line structured
  (`{ before, linkText, after }`) so an edit cannot silently drop the sign-in link.
- Suggestion 1 FOLD: move the announcer and skip link into `+layout.svelte` (or an `Announcer`
  component), and put one `postForm()` in `$lib/api` shared by login, link, and members.
- Suggestion 3 REFUSED: keep the focus ring on programmatically focused headings. It is visible
  focus, and the spec's focus rule does not exempt it.
- Suggestion 4 FOLD: `$state.raw` for `result`, `pending`, and `members`.
- Suggestion 5 FOLD: plain `const` for `refusedParts` and for `+error`'s `siteName`.
- Suggestion 6 ROUTED to rung 7, the same as S11.
- Suggestion 7 FOLD (with B1): Sign out gets a hover state, per the Refinement "Interaction" row.
  The two reviewers disagree on its hover fill (1.3:1 against transparent), so a rendered row
  measures hover and focus contrast on Sign out in both themes.
- Suggestion 8 REFUSED: keep the signed-in header on members `refused` (the re-read's item 12).
  A signed-in non-admin needs sign-out, and a stranger's sign-out is a harmless no-op.
- Suggestion 9 FOLD: with the skip link in the layout, drop the magic `focus:top-20` offset.
- Suggestion 10 FOLD: the roster meta line uses IBM Plex Mono, per the Type table's rule for
  technical metadata.
- Suggestion 11 REFUSED: the deck line covers the case where a name is under 200 characters but
  over 200 bytes.
- Suggestion 12 FOLD: a non-JSON 5xx maps to `failed`, not `network_failure`.
- Suggestion 13 ROUTED to rung 7 (phone gutter `clamp()`).
