# Group-scale principles spec: fold record

Target: `docs/superpowers/specs/2026-09-25-group-scale-principles-design.md`, revision 1
(`1d6bc6a`) to revision 2. Inputs, all under `docs/superpowers/research/`:

- `2026-09-25-group-scale-principles-review-contract.md` (prefix C)
- `2026-09-25-group-scale-principles-review-mechanics.md` (prefix K)
- `2026-09-25-group-scale-principles-review-misuse.md` (prefix U)
- `2026-09-25-group-scale-principles-review-consistency.md` (prefix N)
- `2026-09-25-group-scale-principles-review-prose-early.md` (prefix P; suggestions numbered
  S1 to S9 in file order)

Governing framing (Geoff, 2026-09-25): dubplate should be easy to use for its intended purpose
and hard to use for nefarious ones, and helping people buy music is a bonus. A mechanism that
adds friction for honest users and none for abusers fails that test. Refusals follow
Anthropic's guidance that chasing every finding leads to over-engineering.

## Verification

Each finding was checked against its source before disposition. These load-bearing claims
held:

- Configuration homes "Check out" lets a member copy any server collection they can see
  (`configuration-homes-design.md:498-502`, `:633`), and the group library is an admin-owned
  server collection (`:206`).
- The mirror bypasses per-album intake verdicts (`:621-623`, `:653-654`), copy bypasses intake
  (`:466`), and hook 17 says only intake sets `added_by` (`streamer-2-0-design.md:323-324`,
  `:515-517`).
- The visual identity spec schedules the web album page, with an inscription naming who added
  the album, for rung 7's design review (`2026-09-24-visual-identity-design.md:338-356`).
- `ROADMAP.md` says 0.9's scope is unspecified, and the launch is a 0.x pre-release.
- `config/navidrome.toml` sets neither `EnableSharing` nor `EnableDownloads`.
- Rung 5's add-member page has no removal control and offboarding is an epoch bump
  (`rung-5-identity-auth.md:1099-1103`, `:1445-1447`), so offboarded members keep a roster row.
- Rung 6 moves the roster into the store through `member import` (rung 6 spec, ruling 1).
- The streamer's play-history row carries `client`, not a credential id, and no bytes
  (`streamer-2-0-design.md:299-302`).
- The three rung 5 test names the contract lens cites exist.
- The build spec has an "Out of scope" section and no "principles" section.

Client behavior (Tempus, DSub, Feishin, Symfonium's documentation) and the MusicBrainz sample
are the mechanics lens's external evidence, taken as reported.

## Root fix: the download limit

Five findings converge: K-B1, U-M1, C-M5, U-M2, and P-B4 in part. Revision 1's hard limit
counted album zips, which no offline-sync client in the roster requests. Its "isn't" copy said
members download one album at a time and whole-library copies stay with the operator. Stock
clients falsify both. A limit on per-track fetches would break the offline listening the product
exists for and still meter nothing, because a cache fetch can't be told from playback.

The spec drops the hard limit as its own decision. The framing settles it: the limit added
friction for honest members and none for abusers. In its place, every `stream` and `download`
response is metered per member and per credential in bytes and distinct tracks, and the admin
gets one email when a member crosses a threshold. Nothing is refused. `download` still narrows
to a song or an album. The "isn't" copy now states only what holds. The three lens forks on
this point (K-B1's F1 options b and c, C-M5's fork, U-M1's fork) resolve to "meter plus
notice" under the framing. The one question left, whether a member's own devices holding the
whole library is intended use, is ruling 1.

## Dispositions

### Contract lens (C)

| ID | Sev | Disposition |
|---|---|---|
| C-M1 | major | Folded: new "Acceptance" section, rows A1 to A14. |
| C-M2 | major | Folded: status split into proven, specified, planned, designed-unscheduled, and new; rows re-marked. |
| C-M3 | major | Folded: route-inventory test and table in "Nothing is reachable without a login"; row A2; unauthenticated set restated with the branding, font, and CSS routes. |
| C-M4 | major | Folded: `download` id kinds enumerated, error 70 for the rest, one-album folder directory mapped; row A5. |
| C-M5 | major | Folded with the root fix ("Downloads and offline listening", is/isn't copy). Its fork resolved by the spec as meter plus notice (option b), since option c fails the framing; the residual question is ruling 1. |
| C-M6 | major | Folded with U-M5 and K-m3: one enforcement point, channels, states, CLI raise, bootstrap; row A1. Import folded as a preflight refusal per U-m1 in place of ordered skipping. |
| C-M7 | major | Folded with K-M3 and N-m8: credential id, bytes, dedupe rule; `DisallowUnknownFields` check in A8; forward hook in conflict 4. |
| C-M8 | major | Folded with K-M2, N-M2, K-m7: the per-channel table for `added_by` and the source answer; admin-only edit after admission; intake enforcement; hook beside 17. |
| C-M9 | major | Folded: CLAUDE.md "Refused features" block (conflict 10), build spec "Out of scope" entry (conflict 9), contributor guide with the packaging initiative. Its fork (a subprocess allowlist gate leg) refused: the refusal targets a deliberate feature that review and the written refusals catch, and the test adds upkeep for every honest subprocess while a fork deletes it in one line. Recorded under "Declined mechanisms". |
| C-M10 | major | Folded with N-M1: "Before 0.8" section; sharing pin with a gate leg (A13) and rung 12 carry. Its downloads fork is ruling 3. |
| C-m1 | minor | Folded with K-m1: create, update, and delete answer error 50; `getShares` may answer empty; `shareRole` false; no share route (A4). |
| C-m2 | minor | Folded at the root: the album limit is gone; its store-backed window rule now governs the meter (A6). |
| C-m3 | minor | Folded with K-m5, U-m2, N-m3: the counted set, over-cap behavior, and error 40 on revoke (A7). |
| C-m4 | minor | Folded: album notes written by the `added_by` member or an admin, per the streamer spec. |
| C-m5 | minor | Folded: purchase-type allowlist and `https` only (A14). |
| C-m6 | minor | Folded into the channel table ("Timer path and operator bulk intake"). |
| C-m7 | minor | Folded as a spec decision, not a ruling: the wishlist is web-only, absent from the documented API, and emits no hook event. It costs honest members nothing, so the framing settles it. |

### Mechanics lens (K)

| ID | Sev | Disposition |
|---|---|---|
| K-B1 | blocker | Folded: the root fix. Its F1 fork is ruling 1. Its point 3 (a hard stop's wire shape) is moot, since nothing refuses. |
| K-M1 | major | Folded with N-B1: checkout named in "Downloads and offline listening" and conflict 7, metered like any served bytes. Its F2 fork is ruling 2. |
| K-M2 | major | Folded: the channel table sets `added_by` for the mirror from the credential's member, and hand-over and seize preserve it; conflict 5. |
| K-M3 | major | Folded with C-M7. |
| K-M4 | major | Folded: fallback order (release, sibling releases, labeled artist links) and staleness under "Money points toward buying". |
| K-m1 | minor | Folded with C-m1. |
| K-m2 | minor | Folded with N-1 and P-W2: every share passage listed in conflict 1. |
| K-m3 | minor | Folded with N-6: store-backed roster; every add channel named. |
| K-m4 | minor | Folded differently: the import refuses up front rather than skipping in order (U-m1), so no data is orphaned and no order is needed. |
| K-m5 | minor | Folded: "credential cap", with the reason. |
| K-m6 | minor | Folded with N-10 and P-W4: the answer is taken when each mirrored collection is created. |
| K-m7 | minor | Folded: the source answer is an intake API field. |
| K-m8 | minor | Folded in part: the credential-cap refusal is a web page and never reaches a Subsonic client. The API half is moot, since no API refusal remains. |
| K-m9 | minor | Folded: the view shows counts; the spec says play history records songs and that the member's page says so. |
| K-m10 | minor | Folded with N-4 and P-B2: "designed, unscheduled" status; 0.9 no longer asserted. |
| K-m11 | minor | Folded with N-5: cited as the streamer spec's "2.1 list", not yet re-keyed. |
| K-m12 | minor | Folded: "never fetches" is stated as a project policy, with no shipped fetch recipe. The per-member intake counter refused: members adding music is the intended use, and the admin already sees who added what. |

### Misuse lens (U)

| ID | Sev | Disposition |
|---|---|---|
| U-M1 | major | Folded: the root fix. Its fork resolved by the spec as option a (drop the limit, meter plus notice), under the framing. |
| U-M2 | major | Folded: is/isn't copy restated to what holds, dated to 0.8 where it applies; buy links in the future tense; wishlist dropped from the copy. |
| U-M3 | major | Folded in part: the admin list shows distinct addresses in seven days per credential, and the meter and notice apply per credential. The "most concurrent streams" measure refused: the direct stream path holds no per-stream state, and adding a tracker costs more than a second signal beside distinct addresses and bytes. |
| U-M4 | major | Owner fork: ruling 5, with this lens's option a as the recommendation. |
| U-M5 | major | Folded with C-M6: active members count, admins included, offboarded members don't; the refusal names both ways out; the offboarded state is named as a gap for the enforcing rung. |
| U-m1 | minor | Folded: the Navidrome import refuses before writing when over the cap. |
| U-m2 | minor | Folded: the refusal page lists credentials by last use, marks 90-day-unused first, and offers one-step revoke. |
| U-m3 | minor | Folded: per-upload answer with a per-album override; the mirror answer is a default. |
| U-m4 | minor | Folded: the expanded option list; the "gift" wording on the is/isn't page. |
| U-m5 | minor | Folded: `musicBrainzId` filled from 0.8 (conflict 6). |
| U-m6 | minor | Folded: the source shows beside `added_by` on the album page and in `notes`; the optional store link is captured at upload and becomes the first buy link (ruling 7 states when it renders). |
| U-m7 | minor | Folded as a spec decision, not a ruling: artist-direct first, then mail order, then other stores; streaming links excluded. The framing's pro-artist bonus settles the order. |
| U-m8 | minor | Folded as a note on the unscheduled wishlist, routed to `IDEAS.md` (conflict 11). |
| U-m9 | minor | Folded as a spec decision: filed as an idea in `IDEAS.md` (conflict 11). An unscheduled feature routes to IDEAS by that file's own rule, so this needs no ruling. |

### Consistency lens (N)

| ID | Sev | Disposition |
|---|---|---|
| N-B1 | blocker | Folded with K-M1: checkout is conflict 7; the laptop relay is named under "Servers don't federate". The scope question is ruling 2. The lens's recommendation (owned only) depended on keeping revision 1's "whole-library copies stay with the operator" copy, which the root fix removed as false, so ruling 2 now follows ruling 1. |
| N-M1 | major | Folded: "Before 0.8" section; "login required" re-marked "specified for 0.8"; build spec erratum (conflict 8). Its downloads fork is ruling 3. |
| N-M2 | major | Folded: the channel table covers the timer path, contribute, mirror, copy, hand-over, seize, rung 16, and restore; the "only through intake" line is replaced; no shipped fetching hook recipe. |
| N-M3 | major | Owner fork: ruling 7, re-posed with the rung 7 album page as its premise. The is/isn't copy moved to the future tense. |
| N-1 | minor | Folded with K-m2. |
| N-2 | minor | Folded: the DAP section in conflict 2. |
| N-3 | minor | Folded: one cap over both tables, device credentials per profile, citing the member's existing list. |
| N-4 | minor | Folded with K-m10. |
| N-5 | minor | Folded with K-m11. |
| N-6 | minor | Folded with K-m3. |
| N-7 | minor | Folded: the sensitive-write citation now names "Two surfaces, one API"; the hook 4 citation for stream starts removed. |
| N-8 | minor | Folded with C-M7; conflict 4 says the play-history views filter by event kind. |
| N-9 | minor | Folded: "a consistent layout". |
| N-10 | minor | Folded with K-m6. |
| N-11 | minor | Folded: the wishlist leaves the mechanism table as a spec feature and routes to `IDEAS.md`; one pointer paragraph stays. |
| N-12 | minor | Folded: "get music from the internet"; the "shared shelf" and "distribution node" images removed. |
| N-13 | minor | Folded: the bootstrap is never refused; `member import` is a counted channel; `dubplate init`'s first-run page joins the public table at 0.8. |

### Early prose read (P)

| ID | Sev | Disposition |
|---|---|---|
| P-B1 | blocker | Folded with N-M2: the channel table and "every channel re-validates at the server, and none fetches". |
| P-B2 | blocker | Folded with K-m10. |
| P-B3 | blocker | Folded with N-M1. |
| P-B4 | blocker | Folded with U-M2. |
| P-W1 | warning | Folded: the unsupported claim about scrobbling services removed. |
| P-W2 | warning | Folded with K-m2 and N-2; the conflicts intro no longer calls every entry a conflict. The suggested escape clause for playlist zips refused: a playlist can hold the whole library, and per-track download already serves any client that syncs a playlist. |
| P-W3 | warning | Folded: the build spec's "Out of scope" section is the landing (conflict 9). |
| P-W4 | warning | Folded with K-m6. |
| P-W5 | warning | Folded: "tells the asker". |
| P-W6 | warning | Folded: "such as a family or a few close friends". |
| P-W7 | warning | Folded: plain rewrites for MusicBrainz tagging, "open a collection", and "stream one track at a time". |
| P-W8 | warning | Folded: the setup-colon sentence in the source question split. |
| P-W9 | warning | Folded with N-12. |
| P-W10 | warning | Folded: the determined-operator caveat stated once in the brief and once in "What these principles can and cannot do". |
| P-S1 | suggestion | Folded with N-7. |
| P-S2 | suggestion | Folded: members mint app passwords from a device page. |
| P-S3 | suggestion | Folded: the member cap is sensitive; ruling 4 says each default is a setting, per member where it applies. |
| P-S4 | suggestion | Folded: dates replace "ratified later" in conflict 1. |
| P-S5 | suggestion | Folded: "Money points toward buying". |
| P-S6 | suggestion | Folded: the buy-link row names the first web album page. |
| P-S7 | suggestion | Folded: the repeated "one click" removed from the source ruling. |
| P-S8 | suggestion | Folded: "lossless (FLAC) downloads". |
| P-S9 | suggestion | Refused: triads and bold lead-ins aid scanning in a spec, and tellgrader reports zero tells after the fold. |

## Counts

- Findings: 88 across five inputs (C 17, K 17, U 14, N 17, P 23).
- Folded (wholly or in part), or pointed wholly to a ruling: 87. Several duplicates fold once, listed together above.
- Refused outright: 1 (P-S9).
- Refused parts of folded findings: 5 (C-M9's gate-leg fork, U-M3's concurrency measure,
  K-m12's intake counter, P-W2's playlist-zip escape clause, and K-m8's API half as moot).
- Owner forks: U-M4 and N-M3 point wholly to rulings; the forks inside K-B1, K-M1, N-B1,
  C-M10, and N-M1 feed rulings 1 to 3. Seven rulings total.
- Lens forks the spec settled itself under the framing: C-M5, U-M1, and K-B1's options b and
  c (the hard limit), C-m7 (wishlist scope), U-m7 (link order), U-m9 (local store link).

## Rulings for Geoff

The spec's "Rulings for Geoff" section carries the full text. Revision 1's five rulings map as
follows: its ruling 1 (member cap) and the device half of its ruling 3 are now ruling 4; its
ruling 2 is ruling 6; its ruling 3's album limit is dropped; its ruling 4 is ruling 5; its
ruling 5 is ruling 7.

1. A member's own devices holding the whole library is intended use. Recommendation: yes.
2. Checkout scope. Recommendation: follow ruling 1 (any visible collection if yes, owned only if
   no).
3. Navidrome downloads before 0.8. Recommendation: leave on, state the residual.
4. Defaults. Recommendation: member cap 20, credential cap 10, notice at half the library's
   tracks per member in 24 hours.
5. What the volume view names. Recommendation: members and credentials on bytes and tracks
   only, stream counts as group totals, no time of day, disclosed on each member's page.
6. Source question required. Recommendation: required.
7. When buy links ship. Recommendation: rung 7, if a task absorbs it without growing the pass.

## Owed errata

These are recorded here and in the spec's "Conflicts with existing specs". The fold edited no
ratified document. Each lands after Geoff approves revision 2.

1. `2026-09-07-streamer-2-0-design.md`: drop "Shares" (`:289-295`); rename slice 6 (`:442`);
   remove share management from "Web surface" (`:395`), revision 2 item 12 (`:42`), the API
   launch list (`:384`), and hook 15's share rows (`:507`); narrow `download` (`:384-386`) and
   the "Digital audio players" premise (`:406-409`); add the credential cap and admin list to
   slice 6 and the meter and notice to slice 5; add the serve-event forward hook (credential
   id, bytes, download kind); amend hook 17 and "Added by" (`:323-324`, `:515-517`) to the
   channel table; add the source-answer hook beside 17; add `musicBrainzId` and the source line
   in `notes`.
2. `2026-09-11-configuration-homes-design.md`: retire the reserved share rows (`:564-566`,
   `:675`) and the `createShare` and "shares re-checked at fetch time" entries (`:705-706`);
   record ruling 2's checkout scope (`:498-502`, `:633`).
3. `2026-09-23-rung-6-ops-seam-store.md`: retire the reserved share rows (`:485`).
4. `2026-09-02-greenfield-server-build.md`: the `/share/*` bypass (`:481`) is inert while the
   driver pins sharing off; route the pin and `doctor` row to rung 12; add the refused
   features to "Out of scope" (`:997`).
5. `2026-09-01-dubplate-organizer-design.md`: the 2026-09-12 checkout erratum takes ruling 2's
   scope.
6. `ROADMAP.md`: the checkout line (`:62-63`) takes ruling 2's scope.
7. `CLAUDE.md` (project): a short "Refused features" block (no fetching or fetch
   integrations, no federation, no public sharing, no download larger than an album).
8. `IDEAS.md`: the Requests entry absorbs the wishlist (purchase claim, buy links, web-only
   constraint); a new "our record store" idea.
9. `config/navidrome.toml` and `scripts/check.sh`: pin `EnableSharing = false` with a gate leg
   (a repo change owed by the next code lane, not an erratum to a ratified document).
10. If ruling 3 is "off": `EnableDownloads = false` in the same file, with the same gate leg.

## Refused findings a lens marked blocker or major

None was refused whole. Two major findings had a part refused:

- **C-M9's fork, a subprocess allowlist gate leg.** The refusal it guards targets a deliberate
  feature, which review and the CLAUDE.md block catch. The test adds upkeep for every honest
  subprocess, and a fork deletes it in one line. The rest of C-M9 folded.
- **U-M3's "most concurrent streams" measure.** The direct stream path is stateless, so the
  measure needs a new per-stream tracker. Distinct addresses per credential and the byte meter
  already show a shared credential. The rest of U-M3 folded.

## Second fold

Target: the spec's revision 2 (`08c812f`) to revision 3. Input:
`2026-09-25-group-scale-principles-fold-verification.md` (prefixes B, M, and m; 1 blocker, 3
majors, 14 minors), plus Geoff's rulings of 2026-09-25. The blocker and every major folded.
Minors folded where cheap, and m10, m11, and m14 folded because they touch website copy.

### Owner rulings taken since revision 2

- **Ruling 1, ruled yes.** A member keeping the whole library on their own devices is intended
  use. The meter counts it, the admin sees it, and nothing refuses it.
- **Ruling 2, ruled as its consequence.** A member may check out any collection they can see.
- **Ruling 3, ruled by Geoff's Navidrome answer.** A read-only check found zero shares in
  production and sharing at its default (on). Geoff: "I'm not worried about Navidrome now.
  We'll fix it when we switch it over to dubplate." Sharing and downloads stay at Navidrome's
  defaults until the switch to dubplate's own streamer, as an accepted residual. The fold
  reads "the switch" as 0.8, when dubplate's streamer replaces Navidrome, per the conductor's
  brief. The other reading, rung 16's cutover to dubplate-server with Navidrome still serving
  audio, would move the fix earlier; Geoff can correct it at his read.
- **Shares kept, bounded (direction change).** Casual one-off sharing is an important feature,
  so 0.8 keeps OpenSubsonic shares, bounded so a link is useless as a distribution tool. Geoff
  approved folding "bounded shares" into the spec, with the exact limits set by research that
  runs after this fold. Revision 3 replaces every "drop shares" passage with a placeholder
  section, "Bounded shares (design pending research)", that lists the candidate bounds. Row
  A4 becomes "Shares bounded", with its fixture set by the follow-up. The route inventory
  admits the share link's route as the one public row that serves audio or art. The is/isn't
  copy names a link to one album as intended use and a public library as refused.

### Dispositions

| ID | Sev | Disposition |
|---|---|---|
| B1 | blocker | Folded in part. The facts are stated in "Before 0.8": the 0.63.2 default of `true` (`conf/configuration.go:797`), no key or `ND_ENABLESHARING` in the repo, the tunnel forwarding `/share/*`, and zero shares on 2026-09-25. The false "sets `EnableSharing = false` explicitly" line is gone. The is/isn't copy becomes true by scope: it is website copy for 0.8 onward and publishes when 0.8 ships. The fix's pin, gate leg, and deploy are not taken, because ruling 3 accepts the residual until the switch. |
| M1 | major | Folded, moved to 0.8. "Before 0.8" quotes Navidrome's precedence line and states that any sharing check reads the effective value. A13 becomes the escape hatch's `doctor` row on `getUser`'s `shareRole`, with a fixture whose file says off and whose environment says on, so a file-only check fails. Ruling 3's "Off" option, which had the same shape, is gone with the ruling. |
| M2 | major | Folded. Two rules replace the contradiction. Sources: the first of release, sibling releases, and artist that yields a counted link, with artist links labeled. Order: a host of `bandcamp.com` (or a subdomain) or the artist's `official homepage` host first, then `purchase for mail-order`, then other `purchase for download`, ties by URL. Allowed types named per entity. A14's fixture adds the Bandcamp-host and homepage-host links and the three source cases. |
| M3 | major | Folded. Ruling 5's member-page sentence now discloses the play history whoever runs the server can read, in the verification's wording. The body points to that sentence. The "everything named" option drops "not what you play". |
| m1 | minor | Folded. Distinct tracks are defined over the trailing 24 hours per credential, a member's count is the union over the member's credentials, and a nonzero-offset Range request adds bytes and no track. A6's fixture and pass condition name each case. |
| m2 | minor | Folded. Refused `download` kinds answer error 50, with the OpenSubsonic definitions of 50 and 70 quoted and the reason stated. A5 updated. |
| m3 | minor | Folded. The address source is `serve.ClientIP`; the count shows only in trusted-proxy mode, with a `doctor` row otherwise. A7 covers both modes. |
| m4 | minor | Folded. A fourth auth class, delegated, holds the fallthrough row; A2 names it. |
| m5 | minor | Moot. Geoff ruled ruling 3 directly. |
| m6 | minor | Folded. A7 covers the distinct-address count, new row A15 covers `musicBrainzId`, and A13 covers the sharing `doctor` row. |
| m7 | minor | Folded here. U-M3's refusal of a concurrent-streams measure stands on signal against cost, not on statelessness. The meter's response wrapper is where a counter would sit, so the build cost is small. The signal is weak: a member's phone and laptop streaming at once is ordinary, so the number needs its own threshold and wording, and distinct addresses plus the byte meter already surface a shared credential. |
| m8 | minor | Folded. Ruling 2 is recorded as ruled, as a consequence of ruling 1. Its number stays, because the body and this record cite rulings by number. |
| m9 | minor | Folded. The notice default is half the library or 1,000 tracks, whichever is larger. The evidence is arithmetic: 24 hours of nonstop play at about four minutes a track is about 360 tracks. The repo records no founding-library track count, so the floor carries the case. The spec states that a whole-library sync on a library under 2,000 tracks can pass without an email and still shows in the volume view. |
| m10 | minor | Folded. "The admin can see how much music is downloaded", true under every ruling 5 option. |
| m11 | minor | Folded. "Album pages will link to places to buy the music, where MusicBrainz knows one." |
| m12 | minor | Folded. "Every credited artist carried a buy or stream link, and how many carry a purchase link is unmeasured." |
| m13 | minor | Folded. "Symfonium retries, and its docs tell the user to switch off the download endpoint." |
| m14 | minor | Folded. "What they said about where it came from"; the section publishes at 0.8, so "From 0.8" left the copy; the 122-character line is rewrapped. |

### Owed errata, revised

This supersedes items 1 to 7, 9, and 10 of "Owed errata" above where they differ. The rest
stand.

- Item 1 (streamer spec): the "Shares" erratum becomes "bound shares, per the follow-up". The
  share section, slice 6's name, share management, the launch list, and hook 15's rows stay
  until the follow-up names its changes. The item also records error 50 for refused
  `download` kinds and the escape hatch's `doctor` row on `shareRole` (A13). Its other
  changes stand.
- Item 2 (configuration homes): the reserved share rows and the projector's share tests stay.
  The decline of sharing outside the roster gains the bounded-share exception. The checkout
  half becomes one line saying the meter counts checkout, since ruling 2 confirms the spec's
  scope.
- Item 3 (rung 6 spec): withdrawn. Its reserved share rows stay.
- Item 7 (CLAUDE.md "Refused features"): "no public sharing" becomes "no public library";
  bounded shares are allowed.
- Item 4 (build spec): the `/share/*` bypass is live until 0.8, a residual the owner accepted.
  Any later sharing check reads the effective value. Nothing routes a pin to rung 12. The
  "Out of scope" additions stand, with "public sharing" replaced by "a public library".
- Items 5 and 6 (organizer spec, `ROADMAP.md`): withdrawn. Ruling 2 confirms their checkout
  scope as written.
- Items 9 and 10 (the Navidrome pins and gate legs): withdrawn. No code lane owes a Navidrome
  change.

### Counts

- Findings: 18 (1 blocker, 3 majors, 14 minors).
- Folded wholly: 16 (M1 to M3, and every minor but m5).
- Folded in part: 1 (B1; its pin-and-deploy fix superseded by ruling 3).
- Moot by owner ruling: 1 (m5).
- Refused: 0.
- Rulings: 1 to 3 ruled; 4 to 7 open.
