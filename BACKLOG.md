# BACKLOG

> Individual work items. Managed by `/log-issue`. Initiatives live in `~/Projects/initiatives`; an item that
> belongs to one carries that initiative's id as its tag. The ladder's run order is "Path to 0.8" below.

## Path to 0.8

The order of the remaining work on the greenfield ladder, derived from the dependency analysis
(`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md`), the reverse-edge note
(`docs/superpowers/research/2026-09-04-reverse-edge-completion.md`), the re-plan
(`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`), the intake and `admin-api-1` specs, and the
lean pass process ruling that puts `intake` after `admin-api-1`. A rung is named, and a pass is `<name>-<n>`.

Landed: `foundation`, `corpus`, `matcher`, `shadow` (its program was deleted on 2026-10-05), `auth`, `store`,
`tag-io`, `streamer-1` to `streamer-3`, and `admin-api-1` (the versioned API's cookie mount, member removal
and demotion, the settings write path with its re-authentication gate, the roster verbs, and the members and server
settings pages; the roster API is retired).

1. **`intake`** (rung 7). The uploader and the intake path, as `intake-1` to `intake-3`, ending in Geoff's uploader
   sitting. Inbound: `store`, landed; it followed `admin-api-1` by ruling. `intake-1` (the intake core: the tus
   upload, the gate, the verdicts, and the `/api/v1` submission operations) has landed; `intake-2` is next (the match
   probe, the uploader page, and resume), and it must land before anything deploys, since a passing album waits in
   `matching` with no exit until the probe exists. Carried work: BACKLOG #83 to #85, plus the rest of the items
   tagged `#rung:intake`.
2. **`admin-api-2`** (14b). The second slice of the API, the part that writes a collection row. Inbound: `intake`,
   which creates the default collection. Carried work: BACKLOG items tagged `#rung:admin-api`, the member cap
   (#74) and the collection page's work among them.
3. **`pipeline`** (rung 9). The admission pipeline and the catalog, large enough to split into passes at its own
   planning. Inbound: `store` and `tag-io`, landed; it follows `intake` for the handoff record and the `accepted` state,
   which would otherwise cost a stub producer. It shares no files with `admin-api-2`, so the two may run together.
4. **`streamer`** (rung 12). Slices 3 to 8 of the built-in OpenSubsonic streamer, as `streamer-4` onward (a slice may
   take more than one pass). Inbound: `pipeline`'s catalog. It goes first of the rungs that open after `pipeline`
   because it is the longest and 0.8 ships on it. Slice 7 waits on `queues`. A soft edge ties it to `builtins`, which
   names a post-scan hook event the streamer fires; items 4 to 8 open together, so the order stands.
5. **`queues`** (rung 13). The decision queue. Inbound: `matcher`, `store`, and `pipeline`. It goes ahead of
   `migration` and `builtins` because streamer slice 7 consumes it.
6. **`migration`** (rung 10). Migration and layout parity. Inbound: `corpus`, `store`, and `pipeline`.
7. **`builtins`** (rung 11). The built-in plugin roster. Inbound: `store`, `tag-io`, and `pipeline`; it is
   independent of `migration` and `queues`.
8. **`library-web-1`**, a web pass outside the ladder. The album page, the library grid and list, filtering, and buy
   links, which the intake spec moved to a web pass after `pipeline`. Inbound: `pipeline`'s catalog.
9. **`integration`** (rung 15). The whole-system pass against the launch bar. Inbound: every component rung.
10. **`cutover`** (rung 16). The first launch, 0.8.0, ending in Geoff's first-launch sitting. Inbound: `integration`.
11. **`closeout`** (rung 17). Clean-up after launch. Inbound: `cutover`, and `migration` latest, because the legacy
    ledger fixture is captured before `_salvage/` is deleted.

Items whose edges are met may run at once, up to three passes (ruling 1 of the lean pass process); the numbers set the
start order and are rewritten when the order changes. Items 4 to 8 open together once `pipeline` lands, and items 9 to
11 are strictly serial. Carried work for each item lives in BACKLOG, under items tagged `#rung:<name>`, except where
an item names another home.

Frozen documents keep the old rung numbers, and this table bridges them to the names.

| Old rung | Name | Landed passes (old ids) | New pass ids |
| --- | --- | --- | --- |
| 1 | `foundation` | 1, 1b | none owed |
| 2 | `corpus` | 2a to 2d | none owed |
| 3 | `matcher` | 3a to 3d | none owed |
| 4 | `shadow` | 4a, 4b | none owed |
| 5 | `auth` | 5a to 5d | none owed |
| 6 | `store` | 6a to 6d | none owed |
| 7 | `intake` | none | `intake-1` to `intake-3` (7a to 7c) |
| 8 | `tag-io` | 8a to 8c | none owed |
| 9 | `pipeline` | none | `pipeline-1` onward, split at planning |
| 10 | `migration` | none | `migration-1` onward |
| 11 | `builtins` | none | `builtins-1` onward |
| 12 | `streamer` | slices 1, 2a, 2b | `streamer-1` to `streamer-3` landed; `streamer-4` onward (slices 3 to 8) |
| 13 | `queues` | none | `queues-1` onward |
| 14 | `admin-api` | none | `admin-api-1` (14a), `admin-api-2` (14b) |
| 15 | `integration` | none | `integration-1` onward |
| 16 | `cutover` | none | `cutover-1` onward |
| 17 | `closeout` | none | `closeout-1` onward |
| none | `library-web-1` | none | `library-web-1`, outside the ladder |

## Open

- [ ] #2 Clear the two stale `musicbox-ping-failure@` failed units on the box
  (`systemctl reset-failed`; cosmetic, pre-fix 404 era; box write needs
  sanction) #infra
- [ ] #3 Fix STATUS/runbook wording: musicbox-navidrome was called a "pull"
  check; KV shows all four checks are push (the box's navidrome-check timer
  drives it) #infra
- [ ] #4 pings: `remove-check` has no `--force` escape for a check whose
  route record is permanently missing (undeletable record; lives in the
  pings repo) #infra
- [ ] #5 provision.sh: `ssh_cidrs_json` exits only its substitution subshell
  on malformed input, so the diagnostic degrades to a bare jq error
  #onboarding
- [ ] #6 Verify and close the carried SELinux one-time check on the state
  dir at first deploy (carried from bring-up 2026-08-31; likely moot on the
  running box, needs a stated verdict) #infra
- [ ] #7 Create the Navidrome group accounts (owed; per-contributor intake
  is the uploader's job, but listeners still need logins) #onboarding
- [ ] #8 tus-js-client 5.0 watch item: 5.0.0-pre3 (2026-09-30) exists under the
  `next` dist-tag; re-check at the prod-polish pass and record the upgrade verdict
  #upload
- [ ] #10 provision.sh refuses to run (even `--check`) when
  `FIREWALL_SSH_CIDRS` is unset, but the SSH allowlist was DECLINED
  2026-08-31 (no stable workstation IP), so a fresh provision is impossible
  without setting a variable whose policy answer is "unrestricted".
  Reconcile with a stated decision; do not drive-by edit the guard
  #onboarding
- [ ] #9 Geoff: review the two seed albums parked in
  `/srv/music/review/2026-08-31/geoff/` (Aim "Hinterland (Remastered)",
  illuminati hotties "POWER", plus a TV Girl zine PDF) #pipeline
- [ ] #13 Move the magic-link token from the query (`?t=`) to the URL fragment
  (`#t=`), so a front's access log never records a redeemable token
  (`internal/auth/token.go:181`; the link page's client hook already strips the
  fragment). Changes the link contract the mail, the page, and consumption share.
  The runbook mitigation (drop the query from the access log) landed in 5d-ii.
  From UNI-5d L1, 2026-10-05 #security
- [ ] #14 Align `web/package.json`'s `engines` (`>=22.17`) with the Makefile's web
  target, which refuses a Node major below 24; the runbook follows the Makefile.
  From 5d-ii Task 9A, 2026-10-05 #infra
- [ ] #15 `serve`'s direct-mode warning (9B: a loopback or private peer sent a
  forwarded-address header) fires on `doctor`'s own boundary probe while serve
  runs, so a healthy doctor makes serve log a false alarm. The warning is
  `warnForwardedFromLocalPeer` (`internal/serve/serve.go:205`; its comment opens
  at `:187`). Have the probe's requests (`internal/serve/probe.go`) carry a marker
  the warning skips, or have the warning ignore the probe route. Found at the 5d-ii Task 10 sitting,
  2026-10-05 #bug
- [ ] #16 A username may equal the address of a member who has a username of their own. `principalCollision`
  (`internal/auth/roster.go:830`) already refuses a username equal to an address-principal, but compares
  principals only, so the two principals differ and an audit row can read as that address. Refuse the collision
  on write and at load. Found by the navidrome-removal close,
  2026-10-06. It lands in streamer slice 6 with the `duplicate_principal` code (see #48)
  #security #rung:streamer
- [ ] #18 The rendered gate reaches the add-member username field only through `duplicate_username`
  (`web/scripts/rendered/run.ts`); the client-side pattern refusal is never exercised. Add a row. Found by the
  navidrome-removal close, 2026-10-06 #web
- [ ] #19 The add-member hint text uses `opacity-70` on `base-content`
  (`web/src/routes/members/+page.svelte:323`, and the roster line at `:252`), a contrast the theme gate does not
  measure. Measure it or use a token with a gated pair. Found by the navidrome-removal close, 2026-10-06
  #web
- [ ] #20 No rendered row asserts the autofill-ignore attributes (`autocomplete=off`, `data-1p-ignore`,
  `data-lpignore`) on the add-member Address and Username fields. Add one assertion. Found by the
  navidrome-removal close, 2026-10-06 #web
- [ ] #24 `auth.RunSweep` (`internal/auth/nonce.go:132`) logs a failed sweep's `err.Error()`, which can wrap a
  setting read's text; give it the `SettingError` fixed-reason treatment 6d gave the gate and handlers. Found by the
  6d segment C fold, 2026-10-07 #auth
- [ ] #26 The 6d architecture reads' findings on lines 6d did not touch (unused auth and identity exports, three
  `syncDir` copies, the settings accessor's mutex across a database read, threshold bounds without Floor and
  Ceiling, doctor's file split, and more): `docs/superpowers/research/2026-10-07-6d-architecture-reads.md`, "To
  file". Route to the architecture-cleanup task. 2026-10-07 #arch
- [ ] #28 xmllint 21504 reports a validity error one line too high once the projected document passes line 65535
  (checked with 65536- and 65540-element stdin documents), so `internal/harness/subsonic/xmlrule/check.go`'s
  line-to-ordinal lookup can name the wrong element on a very large body. Guard the size or correct the offset before
  a large recorded session enters the corpus. Found by slice 2b Task 3's review, 2026-10-07 #harness
- [ ] #30 `cmd/harness` cleanup that predates slice 2b, refused at its close as outside the pass. `fetchunidecode.go`
  and `fetchxsd.go` share one cache-fill body. `seeditems.go` and `capture.go` repeat load and capture steps.
  `newCaptureCmd`'s environ and version parameters are a test-only seam. `capture.go:242` and `freeze.go:70` use
  `os.IsNotExist`, which does not unwrap. `capture.go:290` and `synth.go:76` write with `os.WriteFile` where
  `internal/atomicfile` is the house path. Comments in `main.go`, `capture.go`, and `seeditems.go` carry process
  citations. Found by slice 2b's architecture read of `cmd/harness`, 2026-10-10 #harness #arch
- [ ] #31 `stack.Start` should verify the oracle image pin and store the result on the `Oracle`: `harness oracle-up`
  never checks the pin today, and `record-oracle` repeats `Start`'s close-on-failure by hand. Found by slice 2b's
  architecture read of `stack`, 2026-10-10 #harness #arch
- [ ] #32 `fixture`'s `checkPath` misses a file path that is a directory prefix of another (`a/x.flac` and
  `a/x.flac/y.flac`): `Load` accepts the manifest and `Build` then fails at `MkdirAll`. Refuse it at load. Found by
  slice 2b's architecture read of `fixture`, 2026-10-10 #harness
- [ ] #33 For the exported-unused leg's comment-cleanup task: these exports have no caller outside their package
  beyond tests. `grade`'s ledger surface (`State*`, `Class*`, `Pattern`, `ErrInvalidLedger`, `Ledger`, `LoadLedger`,
  `Result`, `Apply`, `Finding`, `Kind*`, `ErrVersionMismatch`) and `stack`'s
  `ErrNoPodman`. Unexport them or name the caller that will use them. Refused at slice 2b's close because the leg is
  staged. `xmlrule`'s five (`ErrVersion`, `ErrSize`, `ErrChecksum`, `ErrNotCached`, `XSDURL`) are unexported (slice 2b Task 12). Found by slice 2b's architecture reads, 2026-10-10 #arch
- [ ] #34 A scoped run that skips the rendered leg launches no browser and could take the light lane, but the lane
  is chosen before `check.sh` classifies the change, so it queues on the heavy lane. Found by the gate scope pass,
  2026-10-10 #gate
- [ ] #35 `internal/harness/cassette/cassette.go` `validate` rejects Extra-vs-Request and Extra-vs-Extra key clashes
  but never two Requests sharing `(Method, PathQuery)`, though the `requestKey` comment says it does; add the
  Request-vs-Request check (from the HISTORY harvest, 2026-10-10) #harness
- [ ] #36 `web/scripts/gates/root-absolute.ts` second `srcset` pattern flags a base64 data URI (every JPEG payload
  starts with `/9j/`) as a root-absolute URL; skip `data:` candidates (from the HISTORY harvest, 2026-10-10) #web
- [ ] #37 The first deploy of the store-backed binary takes the runbook's 6d upgrade ("Upgrade a box that holds a
  roster file"). The same launch work: enabling the serve unit; criterion 14's `runuser` and holder clauses from the
  store spec, with the settings-home pass's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit before it restarts
  `serve`); secrets in every `music`-uid verb's environment; and SMTP from the box. Carried from the alignment pass,
  2026-10-11 #infra #rung:cutover
- [ ] #38 Drive the `systemd-run` sandboxed `member add` and `doctor` on the box, which the store identity pass's smoke never
  exercised, and walk the remaining runbook lines against the live box. Carried from the alignment pass, 2026-10-11
  #infra #rung:cutover
- [ ] #39 Before launch, settle the archive's second copy, per-collection relocation, and store identity (R9's option
  b, whose text stays in the store spec). Carried from the alignment pass, 2026-10-11 #infra #rung:cutover
- [ ] #40 Amperfy's roster walk (`docs/roster-walk.md`) is the attended gate before launch. Tempus's walk ran on
  2026-10-10 (`1f4f232`, 11 of 11 flows). Carried from the alignment pass, 2026-10-11 #streamer #rung:cutover
- [ ] #41 Add a restore check (C14 from the `admin-api` spec review): after a same-box restore, which replaces the
  store and keeps the auth leaf, list each live member whose epoch record is later than the restored snapshot's
  stamp. The runbook already carries the re-run-removals line. Carried from the `admin-api` spec fold, 2026-10-10
  #infra #rung:cutover
- [ ] #42 Streamer slice 3 waits on `pipeline`'s catalog. Its inputs: the XML contract tier from slice 2b, the
  machine-form ledger (52 owed rows, none active), the per-endpoint assignment, the differ, the music-folders
  evidence, and the API base URL measurement the slice 1 study handed on. Carried from the alignment pass,
  2026-10-11 #streamer #rung:streamer
- [ ] #43 Streamer slice 3 also takes UNI-H2, the prefix-gap and percent-encoding items, and the owed ledger rows
  L-20, L-21, and L-51. Carried from the alignment pass, 2026-10-11 #streamer #rung:streamer
- [ ] #44 Update `docs/roster-walk.md` in streamer slice 3 (`streamer-4`): pin Tempus at 4.29.0, not 4.28.0; F11
  expects `createPlaylist(name, songId)`; and the Android emulator needs `-gpu host`, not the doc's `--disable-gpu`
  (`roster-walk.md:35`). Found by the Tempus walk, 2026-10-10 #streamer #rung:streamer
- [ ] #45 The recorded seed lacks lyrics, `nowPlaying`, and radio requests, so the study's expected contract flags
  for them wait for a later recording. Carried from slice 2a, 2026-10-07 #harness #rung:streamer
- [ ] #46 Streamer slice 5 takes the owed ledger row L-17. Carried from the alignment pass, 2026-10-11 #streamer
  #rung:streamer
- [ ] #47 Streamer slice 6 (the revoke verb and app passwords): the store spec's criteria 7 and 8 (the dual-key
  epoch, R-6d20), the `auth` review's minor finding 14, and the store spec's erratum row 10. `EpochStore.Bump` has
  no production caller; whichever of the revoke verb and slice 6's app passwords writes an epoch first weighs
  whether it needs the address key at all. Carried from the store identity pass's routed items, 2026-10-07 #auth
  #rung:streamer
- [ ] #48 Streamer slice 6, the member side: the streamer-admin mapping; the `duplicate_principal` code (with #16);
  offboarding deprovision; CSP scope and shared-origin M2; and username uniqueness. Carried from the alignment pass,
  2026-10-11 #streamer #rung:streamer
- [ ] #49 Streamer slice 7 covers the streamer database in a restore. Carried from the alignment pass, 2026-10-11
  #streamer #rung:streamer
- [ ] #50 Streamer slice 8 takes group 2 in trailers, `CAP_NET_BIND_SERVICE`, and command naming. Carried from the
  alignment pass, 2026-10-11 #streamer #rung:streamer
- [ ] #52 `serve`'s job runner, its instance lease, and job recovery go to the first pass whose page starts a job
  over HTTP (`admin-api` spec `:638`). `admin-api-1` is not that pass: its mount answers the three job operations
  `unavailable`. Carried from the `admin-api` spec fold, 2026-10-10 #ops #rung:admin-api
- [ ] #54 The bearer bypass stays on purpose, and no route registers one yet. Its first consumer (the OpenSubsonic
  API or the bearer admin mount) owes a shell-level cookie test proving a bypassed route never reads the session
  cookie. Carried from the Navidrome removal pass, 2026-10-06 #auth #rung:streamer #rung:admin-api
- [ ] #56 The `id3-tags` reject copy must name a working strip method. The code half landed in `intake-1`
  (`internal/reject` names `mid3v2 --delete-all`); PCF section 4's entry follows at `intake-2`'s copy read (#83).
  Carried from the alignment pass, 2026-10-11 #upload #rung:intake
- [ ] #57 Web carries for `intake-3`, from the `auth` string-deck pass (the intake spec's 7c section points here):
  Ruling 5's four items; W12.4's remaining rules; W4.10's 44px promotion; both palettes' placeholder opacity; the
  outline, dash, and status-ghost text fix (DaisyUI's `--btn-rest-fg` and `--btn-color`), with a text-contrast row
  covering `.btn-dash`; the three latent items from the theme and component-layer pass (the ink border over the outline hover fill,
  a wrapped `<label class="input">` in a status alert, and root-absolute's `srcset` match on a data URI, see #36);
  the favicon paths named twice (`web.go:295-298`, `:485-488`); the styled 404 with the admin gate's refusal through
  the same handler, under a `web-auth-security-reviewer` read, and the error page's onward link (a11y S9); the toggle
  following a later OS scheme change under `system` (a11y S11); a Continue link on the signed-in interstitial and the
  header wordmark's target (N1); the masked "Sign in as a***@" line (security L4); and the phone gutter `clamp()`
  with the interstitials' identity lockup. Carried from the alignment pass, 2026-10-11 #web #rung:intake
- [ ] #58 The uploader takes the theme and component-layer pass's S13 and a11y S7: picked images live at stable paths. Carried from
  the alignment pass, 2026-10-11 #web #rung:intake
- [ ] #59 Run `make derive` once before the first pass that depends on the matcher or the beets oracle (expected:
  `pipeline`), and expect no diff. Carried from the alignment pass, 2026-10-11 #harness #rung:pipeline
- [ ] #60 The catalog's library queries: filter, facets, and trigram search. Carried from the alignment pass,
  2026-10-11 #pipeline #rung:pipeline
- [ ] #61 Before the 0.8 release, recheck TypeScript 7 (7.0.2 when last checked). The bump waits on a svelte-check
  release that supports it: 4.7.6 does not, and `--tsgo` hits rootDir errors (TypeScript #30693,
  `sveltejs/language-tools` #2891). Geoff, 2026-09-26 #web #rung:integration
- [ ] #62 For the project site: settle the `init` and `upgrade` naming against `dubplate-server` (the site spec's
  upgrade naming question), and build the configuration-reference generator as a consumer of forward hook 18's key
  declarations. Carried from the alignment pass, 2026-10-11 #docs (see #100) #initiative:dubplate-server
- [ ] #63 UNI-L2's failure-alert hook for the unidecode table install. Carried from the alignment pass,
  2026-10-11 #infra (see #98) #initiative:dubplate-server
- [ ] #64 The photo checks: `tools/photos/check.test.mjs` fails 2 of 121 (both assume an empty `photos/`), so no gate
  runs it; the photo integrity leg reads `photos/` from disk, not `git ls-files`; and `check.sh`'s header omits that
  leg. Carried from the alignment pass, 2026-10-11 #photos (see #99) #initiative:dubplate-server
- [ ] #65 Photo licensing follow-ups: a source-URL recheck of the pool, and a blob-retention call for withdrawn
  images. Carried from the alignment pass, 2026-10-11 #photos (see #99) #initiative:dubplate-server
- [ ] #66 The architecture cleanup: the routed findings in `docs/superpowers/research/*-architecture-reads.md`
  (5d-ii's doctor twelve-flag OR chain first); the ROUTE items in
  `research/2026-09-27-translit-architecture-reads.md` (the four `TestMain` shim copies, the four copies of the
  owner and leaf checks); #24 to #26; the string-deck pass's five stale-comment nits; the roster outcome classes,
  which no independent oracle cross-checks; and the login title argument and W6's h1 fallback, which have no
  rendered proof. The coverage read's 103 coverage gaps ride along, none blocking. Carried from the alignment pass,
  2026-10-11 #arch
- [ ] #67 The gate's process-citation leg misses `R-6c` patterns (staged legs 317/127). Carried from the alignment
  pass, 2026-10-11 #gate
- [ ] #68 provision.sh: compute the allowed CIDRs once at top level so errexit applies. Today an all-blank
  `FIREWALL_SSH_CIDRS` falls through to jq's failure, and `hcloud firewall create` may get an empty rules file. A
  behavior change; pairs with #5. Found by the alignment pass, PR #2 #onboarding
- [ ] #69 `scripts/check.sh` still says "boundary gate", a retired term. Found by the alignment pass, PR #2 #gate
- [ ] #70 `docs/standing-rules.md` "Gate legs" cites old rung ids and says two legs gate at a comment-cleanup task;
  both legs are warn-only. Found by the alignment pass, PR #2 #gate
- [ ] #83 `intake-2`'s copy read owes the redrafts `intake-1` recorded and did not write. PCF (`passb-copy-final.md`):
  the retry line (`:124-126`) for owner ruling 7; the three Navidrome strings (`:105`, `:143-145`, `:160-161`);
  every "Private Recordings" string in sections 5 and 6, renamed "Unlisted music" with "music" or "recording" where
  "album" would be wrong; the section 1 hint "One album at a time. Albums go in whole." (`:45`); section 4's
  `id3-tags` and extras-warning entries, with the seventeen new reasons `internal/reject` drafted (PCF section 4
  holds only the chip labels: the adopted friendly and next-step strings live in the salvage `reject.go`). Also
  `unsupported-flac`'s wording ("unsupported bit depth ... export it again as 16-bit, 20-bit, or 24-bit"), which
  the gate now also answers for 176.4 kHz and for rates above 768 kHz, where a member told to re-export at a depth
  they already have gets no help,
  `could-not-check`'s "Try again" against the retry path, the "{n} tracks are {label}" alert template that reads
  badly with several labels, and the three `API_DETAILS` lines `intake-1` T10 drafted in `web/src/lib/copy/deck.ts`
  (`store_url`, `extra`, `manifest`). Found by the `intake-1` pass, 2026-10-12 #rung:intake #copy
- [ ] #84 `intake-2`'s planner carries these `intake-1` left out by design: AC 11's two client halves (a dropped
  response ending in the client's success, a lost 201 ending in an upload through `uploadUrl`); AC 23's two race
  clauses through `matching` and `route_unlisted`; AC 26's `route_unlisted` column and its `pushed-back` and
  `match-unavailable` rows; AC 27's `matching` half; the `matching` stall rule and the startup re-queue of `matching`
  (a passing album waits in `matching` with no exit until the probe lands, so `intake-1` must not deploy before
  `intake-2`); `RequireMember`, the page gate, which has no caller until the uploader page; the album's route at
  create, and the separate tracks' entered artist and title (`entered_artist`, `entered_title` are NULL), which land
  with the form. Found by the `intake-1` pass, 2026-10-12 #rung:intake
- [ ] #85 `intake-3` carries the visual identity spec's records the intake spec moved: the catalog line's format
  (VI:243-244), and the list view, filtering, album page, and phone layouts in "Open for rung 7" (VI:350-372), which
  move to the web pass after rung 9; the sitting rules two items. Found by the `intake-1` pass, 2026-10-12
  #rung:intake
- [ ] #86 The streamer spec's forward hooks gain rung 9's MusicBrainz buy-link capture (PS:762), and hook 12 (a
  portable layout) now sits at rung 9. `pipeline` builds both. Found by the `intake-1` pass, 2026-10-12
  #rung:pipeline
- [ ] #87 `intake-2` hardening from `intake-1`'s PR #6 triage (items 1 to 4 and 12): a directory-bytes cap for the ZIP
  walk (`zip.NewReader` allocates each header's name, extra, and comment, up to 192 KB each, after the count passes);
  tusd's own 500 on a filestore fault carries the staging path to the member, which needs a wrapped `ResponseWriter` or
  a datastore wrapper; a backstop that moves a submission to `check-unfinished` after N consecutive verdict-commit
  failures (the one reachable cause, a member named like its archive, is closed); a decode-level `recover` that
  returns `corrupt`, since no panic is known in mewkiz/flac v1.0.14; and `store_url`'s case-insensitive `LIKE`, an
  unbounded `source_note`, manifest work under the store write lock before the `MaxFiles` check, create and cancel
  without a rate limit (the session cap counts only `uploading`), concurrent creates of one file truncating the
  member's own upload, and an authorize-versus-tusd-lock race that lets one PATCH write after a cancel (bounded,
  swept). Found by the `intake-1` pass, 2026-10-12 #rung:intake #security
- [ ] #88 `intake-2` cleanups from PR #6's triage (items 15 and 16): `TestLayering` shells out to `go list` and the
  test cache cannot see it (add it to `UNCACHED_TESTS`); event reasons bypass `reasonReader`; `Limits` and `limits`
  are twin structs; `doctor` re-derives the device comparison instead of calling `staging.Check`; the `collection.go`
  package doc understates the package; dead exports (`reject.All`, `Reason.Blocking`, `Param`, `Canonical`,
  `staging.ErrLinuxOnly`, the `Route` exports) find callers or go; the doctor boundary-probe tests hang to the package
  timeout when `serve` refuses to start; `TestAcquireSharedNonblocking_HeldExclusiveNamesOwnPid` failed once in a gate
  run; `sweepNonces` is never joined at teardown. Found by the `intake-1` pass, 2026-10-12 #rung:intake #test
- [ ] #89 The deploy pass owes, from PR #6's triage (items 9 and 14): confirm the legacy `/srv/music/staging` (the
  musicbox estate's, 2000:2000) is empty or absent before the first deploy, since the first sweep cleans anything in
  it as rowless; narrow `/srv/music` in the serve unit's `ReadWritePaths` with `ReadOnlyPaths` or `InaccessiblePaths`
  for the `music` user's home and legacy directories; `install -d` re-applies owner and mode to an existing library
  root; the state-dir leg's `deploy.sh` check is a substring `grep` (a commented-out line passes); `backup`,
  `restore`, and `member` apply only the lexical staging-versus-library check, and staging is never checked against
  `DUBPLATE_STATE_DIR`. Found by the `intake-1` pass, 2026-10-12 #ops #rung:cutover
- [ ] #90 The route collections' owner is the earliest live admin at creation and is never reassigned, so a demoted
  former admin keeps ownership of "Library" and "Unlisted music"; `admin-api-2` decides what ownership grants and
  whether it follows the roster. Found by the `intake-1` pass, 2026-10-12 #rung:admin-api
- [ ] #91 `pipeline` carries two from PR #6's triage: `already-in-library` can tell an uploader that an album exists in
  a private collection (dedup scope, or a neutral wording); and an album's extras are unchecked bytes of any extension,
  so the rung that first serves extras must set `Content-Type` and `nosniff` or allowlist them by extension. Found by
  the `intake-1` pass, 2026-10-12 #rung:pipeline #security
- [ ] #92 `intake-2`'s first task, three rulings Geoff left open in PR #6's triage (items 5 to 7, disposition "now"):
  keep or drop the minimum-block-size rule in `internal/intake/gate/file.go` (it buys no security, since an attacker
  declares a minimum of 0 or 1, and it can refuse a rare stream-copied file), then add the two missing tests, a clean
  file whose short final frame is below STREAMINFO's minimum and a long silent track under the PCM budget; whether
  `DUBPLATE_INTAKE_MAX_ALBUM_BYTES`, `..._MAX_PICTURE_BYTES`, `..._MAX_TAG_BYTES`, and `..._CHECK_WORKERS` become
  Sensitive (a mark joins only by a ruling); and whether the sign-in endpoint's `DUBPLATE_UNKNOWN_ADDRESS_ANSWER`
  default moves from `kind` to `generic`, since `kind` tells an outsider which addresses are members. Found by the
  `intake-1` pass, 2026-10-12 #rung:intake #security
- [ ] #72 `label.css` and `label-night.css` set no `--size-field`, though the visual identity spec's "Component
  layer" says every theme block sets it. Found by the alignment pass, PR #2 #web
- [ ] #73 Doc corrections: the `internal/serve/serve.go:66-67` typo ("is answers 404"), and the visual identity
  spec's "Copy" section, which still carries the old note about W's section 19. Found by the alignment pass, PR #2
- [ ] #74 The member cap on every add channel (PS:128-151, acceptance row A1; ruling 4's default of 20): one check in
  the roster's add path, `member add` and `add_member` both, the bootstrap never refused, an over-cap roster loaded
  with a `doctor` row. The pass that builds it edits PS:145-147 in place: no command-line set verb exists, and
  `admin-api-1`'s notice rule sends no notice for the seed rule or the command line. The `admin-api-1` plan moved it
  here (its P12), 2026-10-11 #auth #rung:admin-api
- [ ] #75 Removing or demoting an admin sends no notice mail; the spec scopes the notice to sensitive settings, and
  the security read carried it over. From PR #4, 2026-10-12 #auth #rung:admin-api
- [ ] #76 After a restore, an epoch lost with the auth leaf revives a removed member. The runbook's "After a restore"
  step covers it by process until a code fix. From PR #4, 2026-10-12 #auth
- [ ] #77 `api request admitted` is logged before the method and path checks, so an unknown operation logs as
  admitted. Take it in a log pass. From PR #4, 2026-10-12 #ops
- [ ] #78 `doctor`'s address normalizer is a third copy beside `config` and `identity.NormalizeAddress`. From PR #4,
  2026-10-12
- [ ] #79 On a long settings page a successful save shows its alert far from the field; a per-row outcome is a
  redesign. From PR #4, 2026-10-12 #web
- [ ] #80 The settings page's per-row Save buttons are outline until the row's field is dirty, and solid red only on a
  row with an unsaved change (Geoff, 2026-10-11; the visual identity spec reserves solid red for the primary action).
  From PR #4, 2026-10-12 #web
- [ ] #81 The members list marks the signed-in admin's row "(you)" and refuses self-removal; another admin removes
  you (Geoff, 2026-10-11). Today removing yourself signs you out mid-page. From PR #4, 2026-10-12 #auth
  #rung:admin-api
- [ ] #82 Accessibility suggestions the `admin-api-1` close did not take: description ids on the setting inputs, a
  short heading naming the confirm modal, a loading indicator, empty-list lines, a new-tab warning on "Sign in again",
  and human labels beside the environment names. From PR #4, 2026-10-12 #web

## Later

Initiatives that were ROADMAP entries, deferred by Geoff's triage of 2026-10-14. Each takes its pass when the one
before it lands. The entry text below is carried over verbatim from ROADMAP.md as of that triage.

- [ ] #93 The built-in streamer (`streamer`). Subsume Navidrome so the estate is one clean integrated system (Geoff, 2026-09-07): one binary a stranger can set up.
  The spec is `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` (revision 3.1, 2026-10-07; the file still says
  2.0), and it carries the forward hooks and the prior-art list that every rung plan weighs its surface against.
  Navidrome is GPLv3: study it for behavior, never copy. Slices 1 to 2b have landed; next: slice 3, which waits on
  `pipeline`'s catalog. The web player needs a strong play-queue interface (Geoff, 2026-09-24): now playing, up next,
  reorder, and add from album, track, or filter results, designed from a prior-art study in the visual identity spec's
  language, and building on 0.8's saved per-member play queue. The order of the work is "Path to 0.8". #initiative:dubplate-server
- [ ] #94 dubplate-server 0.9: the paired local client (`local-client-0-9`). The 0.9 feature (Geoff, 2026-09-25 ruling on version numbering): a local client
  paired with dubplate-server. It ships the home install, in two modes (Geoff,
  2026-10-10). Standalone: dubplate runs at home on the LAN with no VPS, which Geoff
  expects many home audiophiles to need alone. Home copy: described below. It runs on
  any likely home audiophile setup: native Linux on arm64 and amd64, a multi-arch
  container image (Synology, Unraid, TrueNAS, QNAP), macOS, and Windows. macOS and
  Windows verification is reasoned and disclosed, since the workstation cannot run them.
  0.8 is built ready for it through the streamer spec's forward hooks 23 (no platform
  assumption in the core) and 24 (an install that needs no public host); 0.8 builds
  neither mode. The home copy server (Geoff, 2026-10-10):
  A dubplate install on a machine on a member's home network keeps a one-way synced
  copy of the whole library and serves it to that member's music apps over the same
  OpenSubsonic API. At home the apps stream lossless from the house; away from it they
  use the server. It must run on a current Raspberry Pi (Geoff, 2026-10-10): arm64,
  the library on an attached USB or NVMe drive, and a memory and CPU budget that 0.9's
  plan measures on the then-current model. It builds on 0.8's arm64 builds (Promoted in
  `IDEAS.md`), and it never transcodes, since it serves the stored FLAC untouched. This supersedes the organizer spec's out-of-scope line for "the
  server-to-local mirror" (`2026-09-01-dubplate-organizer-design.md:704`), which stays as
  written. It sits inside the group-scale principles: the copy serves the member's own
  devices, which those principles name as intended use, and it is not a link between two
  catalogs. The `pipeline` plan weighs whether a sync source (a change feed over the catalog,
  keyed on hook 1's audio keys) is a 0.8 forward hook, and routes it to the streamer
  spec's list if so. Everything else, including the pairing credential, the sync
  transport, and how an app picks the home copy or the server, is settled when 0.9 is
  planned.
  The Pi 5 stereo client research (2026-10-10) is the live input for that plan; its current decision is
  `docs/superpowers/research/2026-10-10-pi5-client/proven-build.md`. #initiative:dubplate-server
- [ ] #95 dubplate-server 1.0: the format ladder (`formats-1-0`). (Re-keyed from "1.1"/`formats-1-1`, Geoff, 2026-09-25: the format ladder is 1.0's
  feature; 0.8 and 0.9 stay FLAC-only.)
  Ruled by Geoff, 2026-09-08 (the `tag-io` spec, owner ruling 5, closing fork
  B): 0.8 and 0.9 ship FLAC-only, and the other formats target the 1.0
  release, sequenced after the 0.8.0 launch, which the `streamer` rung carries. The ladder, each
  step its own pass in the shape of `tag-io` (a per-container writer behind
  the seam, a golden container, a corpus stratum derived through beets,
  and a conformance run against the ledger): MP3 (ID3v2.4, the recorded
  version) with Ogg Vorbis and Opus riding the same pass; then M4A (AAC
  and ALAC, one container and one iTunes-atom tag model); then WAV and
  AIFF by transcode to FLAC at intake, which waits on the ffmpeg seam
  and may therefore reorder against 0.8. Everything else stays a reject
  whose message names the supported set and suggests a lossless
  conversion to FLAC where one exists. The `intake` reject copy is written to
  that message shape now, whichever step has landed. Open input: whether
  anyone in the group has ever wanted to upload a non-FLAC file, which decides
  whether the MP3 pass moves ahead of the cutover. #initiative:dubplate-server
- [ ] #96 Give back: permissive Go modules (`give-back`). Geoff, 2026-09-25: where dubplate fills a gap in the Go ecosystem, its reusable layer is
  authored so it can ship as its own permissively licensed module. Each candidate keeps a
  gate-enforced import boundary now (`go list -deps` shows no `github.com/glw907/dubplate/` path)
  and publishes only after a release has proven it. The evidence is in
  `docs/superpowers/research/2026-09-25-give-back-survey.md` and
  `2026-09-25-streamer-prior-art-survey.md`.
  - **OpenSubsonic protocol layer** (decided): hand-written wire types, validated against the
    Apache-2.0 OpenAPI file (reading, 2026-09-25, reversible: the prior-art survey found
    `oapi-codegen` output unfit for production types and `ogen` failing on the full spec, and the
    streamer spec already hand-rolls the dual encoding), a request-parsing handler over a backend interface, a client, and a conformance checker.
    Built with 0.8 and published after 0.8 ships. Clean-room record: behavior cites the OpenSubsonic
    spec or observation, never GPL source.
  - **Matcher** (a beets autotag scoring port, MIT with beets' and `lap`'s notices), together with
    the beets translation half of `internal/engine/musicbrainz` and its conformance suite. Blocker:
    the embedded transliteration table derives from Python Unidecode (GPL-2.0-or-later), so the
    module ships a `func(rune) (string, bool)` seam and the generator, never the table.
  - **FLAC tag I/O** (`tagio`, MIT): a crash-safe Vorbis comment and PICTURE writer on mediafile's
    field table, which no permissive Go library offers. Boundary work: cut `internal/atomicfile`,
    vendor its test fixtures, rewrite process citations.
  - **Conformance corpus** (its own repo): strip MusicBrainz's `tags` and `genres` arrays
    (CC BY-NC-SA) first, and rewrite the README for strangers.
  - Declined: a standalone MusicBrainz client (`musicbrainzws2` covers it), `atomicfile`, the
    store's migrations, and the replay stub.
  - **dubplate stays MIT (Geoff, 2026-09-25; path 1):** the binary stops embedding the
    Unidecode-derived table. It carries a pinned checksum of the Unidecode 1.4.0 package instead;
    at setup it downloads that package, verifies it, and builds the table into the state directory
    with the generator's extraction logic ported to Go, with an offline path from a local file and
    a `doctor` row. The table stays byte-identical, so conformance is unchanged. The matcher reads
    it through the same `func(rune) (string, bool)` seam the published library needs. A small pass,
    due before the first release strangers can install (with `packaging`), plus a lawyer's
    confirmation of the distribution reading before that release. **Landed 2026-10-03**
    (HISTORY, the transliteration entry). Routed here for the module's publication:
    - The lawyer reads the four Unidecode-derived items the repo still carries, not an absolute
      claim: `internal/engine/matcher/testdata/case-mapping-sweep.tsv` (56 rows computed through
      Unidecode 1.4.0), the pinned lookup rows in `internal/engine/matcher/unidecode`'s tests,
      `tools/unidecode-table/generate.py` (imports the GPL library), and `Lookup`'s port of
      `_get_repl_str`'s rules.
    - Git history: the `.tbl` stays in every commit before the pass's Task 2. Whether history is
      rewritten before the repo goes public is the same lawyer's question (the fresh scrubbed
      repo planned at 0.8 may answer it).
    - The transliterator holder is process-wide, installed once by the main package. A published
      matcher library should weigh passing the seam explicitly instead. #initiative:dubplate-server
- [ ] #97 PR test CI (`pr-test-ci`). Add a workflow that runs `bash scripts/check.sh` on `pull_request`. Until it lands, the full-suite home is the
  local close gate (ruling 11, lean pass process spec). It lands after `admin-api-1` (ruling 3, alignment design). #initiative:dubplate-server
- [ ] #98 Packaging (`packaging`). Make the estate easy for a stranger to set up from its ingredients (beets
  or its harvested successor, the binaries, their Cloudflare account, any
  VPS). Opens with a brainstorm; context staged in agent memory (distribution
  model: binary-as-installer vs clone floor; cloud-init generation; existing-box
  bootstrap; pre-setup-VPS option map). **Runs after `dubplate-organizer`**
  (resequenced 2026-09-01, was after `uploader`) so it packages the final
  two-binary shape rather than an intermediate. Five setup pieces moved into 0.8
  (2026-09-24; re-keyed from "2.0", Geoff, 2026-09-25): the TLS listener, the Navidrome import, arm64 builds, `dubplate
  init`, and `dubplate upgrade`; packaging builds on them rather than owning them. The `oss-pair` write-up below
  doubles as its demand test.
  Carried option (Geoff, 2026-09-03): evaluate a Wails desktop app as a
  third head on the shared engine: the existing Svelte UI in each OS's
  own webview (native shell, single small binary per platform; macOS
  notarized DMG, Windows signed installer, Linux Flatpak pinning
  WebKitGTK), runnable embedded (engine in-process, no server) or as a
  client of a hosted estate over the versioned API. Sequencing is firm:
  the bubbletea TUI local organizer builds first after the server; the
  Wails head waits beyond that.
  Configuration brief for the TUI (connection profiles, the machine
  settings and credential files, the attach overlap report and its four
  verbs) and for every settings surface:
  `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`
  (revision 3, 2026-09-12). #initiative:dubplate-server
- [ ] #99 Photo curation for the web surface (`photo-curation`). Curate about 300 openly licensed photographs of music scenes, show flyers, and zine covers for
  the login rotation, the credits gallery, the empty library, and the error pages. The design is
  `docs/superpowers/specs/2026-09-25-photo-curation-design.md` (revision 2.2, approved 2026-09-25),
  which supersedes the visual identity spec's curation rules where they differ; the active plan is
  `docs/superpowers/plans/2026-09-25-photo-curation-round-one.md`. Inputs:
  `docs/superpowers/research/2026-09-24-identity-photos/`. Owner touchpoints: the sorting sittings
  and sending the outreach letters. Separate from the ladder; `intake` consumes it.
  Round one closed on 2026-09-26 with 116 photos committed to `photos/`, and Geoff ruled the pool
  enough for 0.8.0, so later rounds are continuing work that never blocks a release. Each later
  round starts from the targets in `docs/superpowers/research/photo-curation-rounds/round-1.md`
  (Round close) and the artist list beside it, ranks period photos first, and budgets the sorting
  page as its own task. The outreach letters (drafted under the identity-photos research) wait
  until dubplate has a public product to point to (Geoff, 2026-09-26); most of the outreach list's
  gaps, such as early Devo, go-go, and the 1980s dancehall archives, are reachable only that way. #initiative:dubplate-server
- [ ] #100 The project site (`project-site`). dubplate's public project home for strangers: a standard cairn site in its own repo
  (`~/Projects/dubplate-site`) that vendors the theme tokens, the guides, and the generated
  references from a pinned dubplate release, checked against a lock file. Spec:
  `docs/superpowers/specs/2026-09-25-project-site-design.md` (revision 2.5, approved 2026-09-25).
  **Trigger:** cairn's docs reset has closed and `create-cairn-site` is published on npm; the
  build doubles as a stranger's trial of cairn's docs. Phase A builds to a `noindex` preview and
  moves the principles spec's is/isn't lines and legal note into `docs/guide/` (its erratum then
  lands). Two generators are owed to dubplate: the reject-page generator when `intake` merges, and the
  configuration-reference generator when `admin-api` merges. Phase B publishes with 0.8, beside public
  signed binaries and a fresh public source repository from a scrubbed 0.8 snapshot. Open: the
  domain, `dubplate.dev` or `dubplate.studio`, due before Phase B. #initiative:dubplate-server
- [ ] #101 Open-source the pair (`oss-pair`). Publish the uploader and the pings Worker as documentation-linked siblings:
  a validated-intake drop target and a dead-man's-switch monitor, proven here
  first. Preconditions: months of group use, the uploader's auth/validator
  seams, and a clear-eyed maintainership decision. First deliverable,
  independent of any code release: the validated-intake pattern write-up (what
  each layer catches that the others cannot; 907.life is a natural home),
  which costs no maintainership and doubles as the packaging demand test. #initiative:dubplate-server

## Done

- [x] #102 Fix the three errcheck findings golangci-lint 2.15.0 raises on a bare `<-exited` receive:
  `cmd/dubplate-server/backup_test.go:904`, `internal/ops/local/process_test.go:149`, and
  `internal/ops/opshttp/contract_test.go:949`. They fail `make lint` in `scripts/gate.sh` on a
  branch that touches no Go file (found 2026-10-17): each receive now logs the killed child's expected exit error instead of dropping it, 2026-10-17 #infra #initiative:approval-packets
- [x] #55 The first migration after `0005` carried R-6d4's keep limit and joined the older-store table map: `0006`
  adds its tables to `olderstore_test.go`'s map and an undo list for the column and index; R-6d4's keep limit stands as
  6d recorded it, since no measured case asks to change `ServeOptions.Keep`, closed by `intake-1` (T4), 2026-10-12
  #store #rung:intake
- [x] #71 The rung 7 spec's 7c cited HISTORY line 898 for the 5c-iv handoff: repointed to #57, closed by `intake-1`
  (T13), 2026-10-12 #rung:intake
- [x] #12 `member add --help` now marks `--name` and `--username` "(required)": closed by the `admin-api-1` pass
  (T7), 2026-10-11 #security
- [x] #17 The members page maps every add refusal detail, `invalid_address` included, to its field for the invalid
  state and the focus (`FIELD_FOR_DETAIL`): `admin-api-1` T9, 2026-10-11 #web
- [x] #21 `deck-coverage.ts` reads its one Go outcome source and checks only that the parse is non-empty and names
  `signed_out`, with no exact count: `admin-api-1` T11, 2026-10-11 #web
- [x] #23 The roster API is retired and the pages read `/api/v1`'s JSON refusals first, so a 403 no longer shows the
  cross-site copy: `admin-api-1` T9 and T11, 2026-10-11 #web
- [x] #25 The seed rule wraps the parse error, never a nil `Validate`: `admin-api-1` T4, 2026-10-11 #config
- [x] #51 The admin write path: the last-admin guard, removal with the epoch bump and demotion on the actor's
  `UserID`, and `settings reset` for an admin row that refuses `serve`'s start: `admin-api-1` T3 to T8, 2026-10-11
  #auth
- [x] #53 The roster line's blank-username conditionals, the unreachable `duplicate_principal` mapping, and the
  fixture-roster comments are gone with the roster API: `admin-api-1` T9 and T11, 2026-10-11 #web
- [x] #1 Write the Go-rewrite retrospective: closed under lean ruling 8: history is git and PR bodies, 2026-10-11
  #infra

- [x] #11 Resolve the docker compose leg's `command -v docker` skip: retired with the leg
  itself (`d1e93e9` removed the compose leg from `scripts/check.sh`), 2026-10-05 #infra
- [x] #29 The session credential scan treated every string leaf holding `?` as a URL query: it now splits the leaf
  on `?`, `&` and `;`, reads only `key=value` fields, and refuses nothing for prose (`be9abf5`), 2026-10-10 #harness
- [x] #22 `harness capture` overwrote an existing case with no existence check: capture and synth now share one
  overwrite guard (`e561c32`, slice 2b Task 1), 2026-10-10 #harness
- [x] #27 `check.sh --scope-changed` ran the rendered leg for a harness-only change: the conform and rendered scopes
  are derived apart (`e561c32`, slice 2b Task 1), 2026-10-10 #gate
