# Rung 5c-iv register review

This record folds the register gate's `prose-voice-reviewer` findings on
`web/src/lib/copy/deck.ts` (Task 7, `e4ce78b`) into the string deck. Each item below carries one
disposition and the evidence checked against the Go source before the line changed.

**Counts:** 25 items. 24 taken as findings against the deck, `title.ts`, `deck-coverage.ts`, or
the 5c-iii plan text; 1 (item 23) recorded for the owner, no change made.

**Follow-up fold.** The dispatch that produced `e4ce78b` omitted the reviewer's exact rewrite
text for several items, so those items landed marked "taken" or "left unchanged" without the
line actually changing in `deck.ts`. A later pass applied the omitted rewrites and updated the
affected items below (parts of item 5, all of 8-13 except `signed_out`/`rate_limited`, and all
of 16-22 except `ERROR_PAGE.genericHeading`) to "folded," each verified against the same Go
citations already recorded in "What this review verified before acting."

## Dispositions

1. **`ADD_MEMBER_HINTS.streamerUsername` (`deck.ts:94`).** Taken. `internal/auth/roster.go:43-52`
   (`streamerUsernamePattern`) requires an alphanumeric first character, then any run of
   `[A-Za-z0-9._@+-]`, and `roster.go:37-41` bounds it at 64 bytes. The pattern is ASCII-only,
   so "characters" and "bytes" agree; the hint now states the shape and the count.
2. **`ROSTER_OUTCOMES.write_unconfirmed` (`deck.ts:113`, now `:122-123`).** Taken verbatim, per
   `internal/auth/admin.go:322-326` (`errRosterWriteLanded` → "the member may be on disk").
3. **`duplicate_principal` (`deck.ts:105`, now `:114`).** Taken verbatim.
   `internal/auth/roster.go:111-118, 828-839` (`principalCollision`): since Task 9a's form now
   requires `streamer_username` on every add (see item 15), an add-time `errDuplicatePrincipal`
   always means the entered streamer username collides with another member's principal where
   that other member has no streamer username of its own, i.e., the entered username equals
   another member's address.
4. **`AUTH_OUTCOMES.requested` (`deck.ts:38`).** Taken verbatim.
   `internal/auth/handlers.go` answers `requested` identically whether or not the address is on
   the roster (the generic, non-enumerating answer), so the line describes that behavior rather
   than claiming a specific address status.
5. **Drop "roster" from member-facing lines.** Folded for `added`, `duplicate_address`, `busy`,
   and `unknown_field` (W8.2's jargon-leak catalogue lists "roster"). Text: `Added.`; `That
   email address already belongs to a member.`; `That didn't go through. Try again.`;
   `unknown_field`'s final text is folded again under item 16-22 below, since the first fold's
   `unknown_field` wording did not survive the follow-up rewrite. (Apostrophes are U+2019 in the
   source; ASCII shown here.)
6. **`unknown` and `unknownAddressLine` (`deck.ts:26`, `:60`, now `:26`, `:59-63`).** Taken.
   Both lines now suggest checking the spelling before naming the admin contact.
   `unknownAddressLine`'s signature changed from `(adminContact: string | null)` to
   `(adminContact: { displayName: string; address: string } | null)`, matching
   `internal/auth/handlers.go`'s `adminContact` struct (`DisplayName`, `Address`,
   `json:"display_name"`/`"address"`) at `:163-166`. Its one caller,
   `web/scripts/gates/deck-coverage.ts`'s `deckLeaves()`, is updated to pass the object shape.
   No page yet calls it (Task 8 is not built).
7. **`refused` (`deck.ts:46`).** Taken verbatim. Added a note to the 5c-iii plan's Task 8 text
   (the interstitial paragraph) that the `refused` copy links back to the sign-in page, since
   the deck line now says "Get a new one from the sign-in page."
8-13. **`invalid_display_name`, `invalid_streamer_username`, `failed`, `rate_limited`,
   `CROSS_ORIGIN_REFUSED`, `signed_out`.** `signed_out` and `rate_limited` taken with the given
   text. `signed_out`: "You're signed out in this browser." (was "on this device"; a
   magic-link server clears one cookie in one browser, which is more precise than "device").
   `rate_limited`: "Too many requests. Wait a while and try again." — `internal/auth/limit.go`'s
   `rateFloor = 1` (`cmd/dubplate-server/auth.go:143`) and `limitInterval = time.Hour`
   (`limit.go:22`) mean a limit configured at the floor refills only once per hour, so "a few
   minutes" is untrue at that configuration; "a while" makes no minute claim. The other four are
   folded with the reviewer's exact rewrite text, verified against the Go source before writing:
   `invalid_display_name`: "That name is too long, or it has characters it can't use, such as
   tabs or line breaks." — `internal/auth/roster.go:580` calls `displayNameCarriesRefusedRune`
   (`:673-681`, every control/format character and every whitespace rune but the ordinary space)
   and the 200-byte bound (`maxDisplayNameLength`, `roster.go:35`); the line says "characters,"
   not a byte count, since the field is not ASCII-only the way `streamer_username` is.
   `invalid_streamer_username`: "Start the username with a letter or number, and use only
   letters, numbers, periods, underscores, hyphens, plus signs, or @." — matches the hint text
   already folded in item 1 and `streamerUsernamePattern` (`roster.go:52`,
   `^[A-Za-z0-9][A-Za-z0-9._@+-]*$`). `failed`: "Something went wrong on our end. Try again in a
   few minutes." — every `outcomeFailed` write site is a server-side condition
   (`internal/auth/handlers.go:221`, roster unavailable; `:258-259`, admin-contact lookup
   failure; `:271-272`, unmapped send failure; `internal/auth/admin.go:207`, roster listing
   failure; `:326-327`, `additionOutcome`'s default case for an unmapped write error), never a
   client mistake,
   so "on our end" is accurate and "in a few minutes" replaces the vaguer "Try again."
   `CROSS_ORIGIN_REFUSED`: "We couldn't accept that from another site. Go to this site's address
   directly and try again." — names the fix (browse to the site directly) instead of only
   naming the refusal.
14. **`TITLES.notFound` and `TITLES.error`.** Taken. Added both keys; `+error.svelte` now reads
   `TITLES.notFound`/`TITLES.error` and `ERROR_PAGE.notFoundHeading`/`genericHeading` instead of
   its stub literals, composing the `<title>` through `pageTitle`. `deck-coverage.ts`'s
   `REQUIRED_TITLES` gained both keys, and gained the reverse check (every `TITLES` key must
   appear in `REQUIRED_TITLES`), so the two lists cannot drift apart in either direction. The
   5c-iii plan's Task 9a text now names `TITLES.notFound` for the non-admin members state
   (`internal/auth/admin.go:96-100`'s `RequireAdmin`: a non-admin's request answers exactly
   `http.NotFound`, carrying no cookie, so the tab must not say "Members" to it).
15. **Task 9a's "an optional streamer username."** Taken as a plan-text correction, not a
   code change (the write path was already correct). `internal/auth/roster.go:606-615`
   (`validateWriteCandidate`) refuses an empty `streamer_username` with
   `errMissingStreamerUsername`, so the field is required on every add, not optional. The plan's
   outcome bullet and its "2.0 hook" line (which cited R-5c13 for the optionality) are both
   corrected: R-5c13 governs only the admin field's absence from the form
   (`docs/plans/2026-09-23-rung-5c-pages-member-verb.md:731`), not the streamer
   username.
16-22. **`too_large`, `malformed`, `unknown_field`, `admin_not_grantable`, `invalid_address`,
   `duplicate_streamer_username`, `INTERSTITIAL_STATEMENT`, `ERROR_PAGE.genericHeading`.**
   The first seven are folded with the reviewer's exact rewrite text, each naming what the
   member should do next instead of only naming the problem: `too_large`: "That was too much to
   send. Check what you entered and try again." `malformed`: "Something went wrong sending that.
   Reload the page and try again." `unknown_field`: "This page sent something the server doesn't
   accept. Reload the page and try again." (supersedes item 5's earlier "That form couldn't be
   read. Try again."; both drop "roster," and the reload instruction is more actionable).
   `admin_not_grantable`: "Admins can't be added from this page." — `formRefusal`
   (`internal/auth/admin.go:271-295`) refuses `admin_not_grantable` whenever the submitted form
   carries the admin field at all, in the body or the query (`:282-284`), so the outcome is
   always "the form named an admin field," never a demoted-actor or write-time refusal; the
   5c-iii add-member form carries no admin control (item 15's note), so "from this page" is
   accurate.
   `invalid_address`: "That doesn't look like an email address. Check it and try again."
   `duplicate_streamer_username`: "That streamer username is already taken. Choose a different
   one." `INTERSTITIAL_STATEMENT`: "To finish signing in, select Sign in." —
   `INTERSTITIAL_CONTROL` (`deck.ts`) is "Sign in," so the statement names the exact control
   label per the rewrite's own condition. `ERROR_PAGE.genericHeading` has no body-line key in
   the deck (only `notFoundHeading` and `genericHeading`, both consumed by `+error.svelte`), so
   left unchanged; no key was added.
   `SENT_MOMENT_TEXT` now reads `AUTH_OUTCOMES.sent` instead of repeating the literal, so the
   sent moment's text and the `sent` outcome's line cannot drift. No separate sent-moment
   heading key was added: the existing `SENT_MOMENT_TEXT` already serves as the moment's
   heading text per Task 8's design (the moment is a single text block; there is no second,
   distinct heading string to name), so no new key was needed.
23. **"Streamer username" as a term.** Left as is, recorded for the owner: the reviewer flagged
   it as a possible jargon leak (a term a member must learn to use the built-in streamer), but
   changing it touches every label, hint, and outcome line that names the field and has no
   drop-in replacement that stays as precise. No change made; this is the owner's call.
24. **The false comment at `deck.ts:47-49` (now `:48-49`).** Taken. Replaced "this line
   describes only this device's session ending, since a magic-link server has no way to end a
   session on any other device" with "sign-out clears only this browser's cookie and revokes no
   session," matching the corrected `signed_out` line's own wording (item 13).
25. **`title.ts:14`'s dangling separator.** Taken. `pageTitle(state, siteName)` now returns
   `state` alone when `siteName` is empty, instead of `"${state} · "`. Added
   `checkPageTitleEmptySiteName` to `deck-coverage.ts`'s gate, asserting
   `pageTitle(TITLES.login, '') === TITLES.login`.

## What this review verified before acting

- `internal/auth/roster.go:43-52` (the pattern), `:37-41` (the byte bound), `:606-615`
  (`validateWriteCandidate`, the required-field check), `:111-118` and `:828-839`
  (`principalCollision`), `:1880-1950` (`roster_test.go`'s load/write collision tests, showing
  every write-time `errDuplicatePrincipal` test case is a username colliding with an address,
  never the reverse, consistent with the field now being required on every add).
- `internal/auth/admin.go:280-326` (`additionOutcome`, mapping every roster write error to its
  outcome, including `outcomeWriteUnconfirmed` at `errRosterWriteLanded`) and `:35-46` (the
  outcome constant declarations `deck-coverage.ts` parses).
- `internal/auth/handlers.go:140-166` (the `answer` and `adminContact` JSON shapes).
- `internal/auth/limit.go:19-22` (`limitInterval = time.Hour`) and
  `cmd/dubplate-server/auth.go:139-156` (`rateFloor = 1`, the four rate ceilings), confirming a
  floor-configured limit refills hourly, not "in a few minutes."
- `docs/plans/2026-09-23-rung-5c-pages-member-verb.md:731` (R-5c13's actual scope:
  the admin field's absence from the form).
- `web/src/routes/+error.svelte`, `login/+page.svelte`, `login/link/+page.svelte`,
  `members/+page.svelte`: confirmed by `grep` that none of these pages import `$lib/copy/deck`
  or `$lib/copy/title` yet (Task 8/9a/9b are unbuilt structural stubs), so `+error.svelte` is
  this pass's only existing reader to update; the others are out of this task's scope.
- `npx svelte-check --threshold error` (308 files, 0 errors) after editing `+error.svelte`.
- `node web/scripts/gates/deck-coverage.ts` (53 deck values checked, passed) and
  `node web/scripts/gates/ai-tells.ts` (60 scanned strings, passed) run directly before the full
  gate.

**Follow-up fold's own verification (items 5, 8-13, 16-22):**

- `internal/auth/handlers.go:221`, `:258-259`, `:271-272` and `internal/auth/admin.go:207`,
  `:326-327`: every `outcomeFailed` write site, confirming `failed` is always server-side.
- `internal/auth/roster.go:575-582` (`validate`, the `errInvalidDisplayName` check) and `:673-681`
  (`displayNameCarriesRefusedRune`), plus the `maxDisplayNameLength = 200` declaration
  (`roster.go:35`), for `invalid_display_name`.
- `internal/auth/admin.go:271-295` (`formRefusal`, including the admin-field check at `:282-284`
  and the unknown-field check at `:286-289`), for `admin_not_grantable` and confirming
  `unknown_field`'s Go-side meaning is unchanged from the earlier fold.
- `web/src/lib/copy/deck.ts`'s own `INTERSTITIAL_CONTROL` value ("Sign in"), for
  `INTERSTITIAL_STATEMENT`.
- `web/src/routes/+error.svelte` and `deck.ts`'s `ERROR_PAGE` shape, confirming no body-line key
  exists for `genericHeading` (so it was left unchanged, per the task's own condition).
- `node web/scripts/gates/deck-coverage.ts` (53 deck values checked, passed) and
  `node web/scripts/gates/ai-tells.ts` (60 scanned strings, passed), rerun after this fold.
