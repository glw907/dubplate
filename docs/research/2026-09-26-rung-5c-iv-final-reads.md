# Rung 5c-iv final read and architecture reads, and fold R3 (2026-09-26)

These are the last fresh-context capture read (at `aa8e19e`) and Task 11's two architecture
reads, with the conductor's rulings. Fold R3 takes every item marked FOLD. Everything marked
ROUTE goes to the architecture-cleanup list, or to the pass named.

## Final capture read (Opus, 60 native cells and 15 contact sheets): fold

- D1 FOLD: at 320px the signed-in header's wordmark touches "Sign out" in every members state,
  in both schemes (the "e" ends at x=207 and the "S" starts at x=208). The cause is
  `SiteHeader.svelte:96-103`, where `navbar-start` and `navbar-end` each take half the row and
  the row declares no breakpoint. This violates W4.4 and W4.1. Fix: the row declares its own
  breakpoint, with the smallest change that clears the collision (a size and gap step below
  it). Showing the mark alone is rung 7's open item, so use it only if nothing smaller works,
  and report which fix landed. A rendered row asserts at least 8px between the two boxes at
  320 and 390.
- C1 FOLD: the member meta line is `font-mono text-sm` (14px, `members/+page.svelte:275`). The
  spec allows mono at 11, 12, or 13px only. It is a catalog-style line, so use 12px.
- P1 FOLD: login-sent repeats "Check your email" in the h2 and the line under it. Give the line
  its own words in the deck.
- P2 FOLD: login-sent's h2 sits 11px above and 4px below the label rule. W3.9 asks for large
  space above a heading and tight space below.
- P3 FOLD: the meta line wraps with a leading " · ". Keep the separator attached to the segment
  before it, per the catalog-line discipline.
- P4 FOLD: "Added." names no one. Make it a deck function that names the added member.
- N1 ROUTE to rung 7: the header's wordmark links to /members, which is not found for a
  non-admin. This routes with W8's Continue, until a member page exists.

## Architecture reads

## web/ at aa8e19e: workmanlike (src density 0.40 vs kit runtime/client 0.17; excess is ~50 process-citation lines)
Correctness, FOLD in R3:
- SiteHeader.svelte:87: every failed sign-out except cross_origin announces NETWORK_FAILURE, including a 5xx; run.ts:4093-4096 pins the wrong line. A non-2xx answer reads AUTH_OUTCOMES.failed. (The conductor's R1 ruling text named only the two lines; the fault is the ruling's wording.)
- Root: outcome-to-text is written three times (auth.ts:45-88; members/+page.svelte:118-143; SiteHeader.svelte:87). One $lib/api/outcome.ts resolver over both dictionaries, with one class table; members and the header call it.
- members/+page.svelte:167 unchecked cast of the roster JSON; {} or {members:null} throws at :266. Shape-check as post.ts:51 does.
- run.ts:5032 waits for .btn:disabled, which the app never sets (aria-disabled everywhere): the pending guard can never fire. Match .btn[aria-disabled="true"].
- run.ts:2473-2480 independent oracle omits signed_out, which auth.ts:63 classes as a status.
- deck-coverage.ts:105-133 lists deck exports by hand, while deck.ts:13-15 promises every export is covered; walk import * as deck. TITLES as const.
Architecture, ROUTE to cleanup: roster API client inside the members route (-> $lib/api/roster.ts); unused exports (deck.ts:84 SENT_MOMENT_TEXT, :25, announcer Politeness, auth AdminContact; many script gate exports); run.ts 5,248 lines -> split by concern, rename process-named rows; gates named by spec number, shared parsers inside gates; WCAG luminance twice (g9.ts:155-212 at 0.03928, run.ts:1386-1400 at 0.04045); #outcome id is a test hook; members Field union x4, NotFound component; login uses ADD_MEMBER_LABELS.address; deck dictionaries typed Record<string,string> (use as const satisfies); ~50 citation lines; browser ? guards unreachable under ssr=false.
Best: hooks.client.ts + link-token.ts; announcer.svelte.ts; post.ts:41-54.

## cmd/dubplate-server at aa8e19e: workmanlike (density 0.429 vs cmd/gofmt 0.298; rendered_test.go 0.260)
Since e90998a: storeVouches closes the prior latent finding; the prior read's findings stand.
- Harness delivers links two ways: seven pre-minted DUBPLATE_RENDERED_*_LINK env vars (rendered_test.go:829-851, :858-876; run.ts:5173-5180) and mintServer per cell. Make mintServer the only path; drop renderedMagicLinkExpiry (:48-53), its override (:394), the guard (:811-813), and the Makefile RENDERED_TIMEOUT coupling (Makefile:106-112). ROUTE: architecture cleanup (harness).
- requirePOST (:563-572, :586, :632) restates serve.go:378-389 methodOnly and drops Allow; wrap with methodOnly. ROUTE: cleanup.
- rosterResetServer's baseline is whatever the file holds at the first POST (:613-659), valid only by script order; make snapshot/reset explicit (409 before snapshot) or let Go own the baseline. ROUTE: 6d Task 4, which rewrites this server onto the users table.
- Literal env key names in the harness (:540, :545, :855, :919; map :382-395) duplicate serve.go:26-31 constants and threshold .Key; mintLinkFor (:535-550) re-resolves auth config per mint. ROUTE: cleanup (6d Tasks 6-7 remove resolveAuthConfig anyway).
- checkRateBudget and checkSpendReconciliation duplicate a sum loop (:495-499, :681-685); slices.Concat at :89. ROUTE: cleanup.
- backup.go:360-364 three-line comment -> one-line label or own case. ROUTE: cleanup.
- New process citations rendered_test.go:232, :577, :66, :443, :475; renderedScriptTimeout doc (:35-41) is history; :241 over-long. ROUTE: comment cleanup.
Best: store.go:187-197 storeVouches; rendered_test.go:673-702 with :289-294 spend reconciliation; :730-750 scriptEnvNames allowlist.

## Fold R3's outcome and R3b (2026-09-26)

R3 landed in two commits, `4c5864a` (its subject mislabels it "Task R2") and `8337877`. Its
reviewer escalated one literal gap. `post.ts` still mapped a non-JSON 4xx other than 403 to
`network_failure`, so a sign-out answered by a 404 page announced the connection line.

**Ruling (conductor):** an answer that arrived means the server was reached. So every arrived
non-JSON answer reads `failed`, a 403 still reads `cross_origin`, and `network_failure` is only
for a fetch that threw. This takes path (a), which changes the shared mapping for login and link
as well. R3b folds it, with three of the reviewer's cheap non-blocking items:

- the `requested` line's repeated "Check your email";
- deck-coverage failing on a nested object export;
- `memberCountLine(1)` sampled.

**Routed to cleanup:** no independent oracle cross-checks the roster outcome classes.
