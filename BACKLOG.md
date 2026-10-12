# BACKLOG

> Individual work items. Managed by `/log-issue`. Initiatives in `ROADMAP.md`.
> Live-use findings during the Go development land here with
> `#project:prod-polish` so the cleanup pass collects them.

## Open

- [ ] #2 Clear the two stale `musicbox-ping-failure@` failed units on the box
  (`systemctl reset-failed`; cosmetic, pre-fix 404 era; box write needs
  sanction) #infra #project:prod-polish
- [ ] #3 Fix STATUS/runbook wording: musicbox-navidrome was called a "pull"
  check; KV shows all four checks are push (the box's navidrome-check timer
  drives it) #infra #project:prod-polish
- [ ] #4 pings: `remove-check` has no `--force` escape for a check whose
  route record is permanently missing (undeletable record; lives in the
  pings repo) #infra #project:prod-polish
- [ ] #5 provision.sh: `ssh_cidrs_json` exits only its substitution subshell
  on malformed input, so the diagnostic degrades to a bare jq error
  #onboarding #project:prod-polish
- [ ] #6 Verify and close the carried SELinux one-time check on the state
  dir at first deploy (carried from bring-up 2026-08-31; likely moot on the
  running box, needs a stated verdict) #infra #project:prod-polish
- [ ] #7 Create the Navidrome group accounts (owed; per-contributor intake
  is the uploader's job, but listeners still need logins) #onboarding
  #project:prod-polish
- [ ] #8 tus-js-client 5.0 watch item: 5.0.0-pre3 (2026-09-30) exists under the
  `next` dist-tag; re-check at the prod-polish pass and record the upgrade verdict
  #upload
- [ ] #10 provision.sh refuses to run (even `--check`) when
  `FIREWALL_SSH_CIDRS` is unset, but the SSH allowlist was DECLINED
  2026-08-31 (no stable workstation IP), so a fresh provision is impossible
  without setting a variable whose policy answer is "unrestricted".
  Reconcile with a stated decision; do not drive-by edit the guard
  #onboarding #project:prod-polish
- [ ] #9 Geoff: review the two seed albums parked in
  `/srv/music/review/2026-08-31/geoff/` (Aim "Hinterland (Remastered)",
  illuminati hotties "POWER", plus a TV Girl zine PDF) #pipeline
- [ ] #13 Move the magic-link token from the query (`?t=`) to the URL fragment
  (`#t=`), so a front's access log never records a redeemable token
  (`internal/auth/token.go:181`; the link page's client hook already strips the
  fragment). Changes the link contract the mail, the page, and consumption share.
  The runbook mitigation (drop the query from the access log) landed in 5d-ii.
  From UNI-5d L1, 2026-10-05 #security #project:prod-polish
- [ ] #14 Align `web/package.json`'s `engines` (`>=22.17`) with the Makefile's web
  target, which refuses a Node major below 24; the runbook follows the Makefile.
  From 5d-ii Task 9A, 2026-10-05 #infra #project:prod-polish
- [ ] #15 `serve`'s direct-mode warning (9B: a loopback or private peer sent a
  forwarded-address header) fires on `doctor`'s own boundary probe while serve
  runs, so a healthy doctor makes serve log a false alarm. The warning is
  `warnForwardedFromLocalPeer` (`internal/serve/serve.go:205`; its comment opens
  at `:187`). Have the probe's requests (`internal/serve/probe.go`) carry a marker
  the warning skips, or have the warning ignore the probe route. Found at the 5d-ii Task 10 sitting,
  2026-10-05 #bug #project:prod-polish
- [ ] #16 A username may equal the address of a member who has a username of their own. `principalCollision`
  (`internal/auth/roster.go:830`) already refuses a username equal to an address-principal, but compares
  principals only, so the two principals differ and an audit row can read as that address. Refuse the collision
  on write and at load. Found by the navidrome-removal close,
  2026-10-06. It lands in streamer slice 6 with the `duplicate_principal` code (see #48)
  #security #project:prod-polish #rung:streamer
- [ ] #18 The rendered gate reaches the add-member username field only through `duplicate_username`
  (`web/scripts/rendered/run.ts`); the client-side pattern refusal is never exercised. Add a row. Found by the
  navidrome-removal close, 2026-10-06 #web #project:prod-polish
- [ ] #19 The add-member hint text uses `opacity-70` on `base-content`
  (`web/src/routes/members/+page.svelte:323`, and the roster line at `:252`), a contrast the theme gate does not
  measure. Measure it or use a token with a gated pair. Found by the navidrome-removal close, 2026-10-06
  #web #project:prod-polish
- [ ] #20 No rendered row asserts the autofill-ignore attributes (`autocomplete=off`, `data-1p-ignore`,
  `data-lpignore`) on the add-member Address and Username fields. Add one assertion. Found by the
  navidrome-removal close, 2026-10-06 #web #project:prod-polish
- [ ] #24 `auth.RunSweep` (`internal/auth/nonce.go:132`) logs a failed sweep's `err.Error()`, which can wrap a
  setting read's text; give it the `SettingError` fixed-reason treatment 6d gave the gate and handlers. Found by the
  6d segment C fold, 2026-10-07 #auth #project:prod-polish
- [ ] #26 The 6d architecture reads' findings on lines 6d did not touch (unused auth and identity exports, three
  `syncDir` copies, the settings accessor's mutex across a database read, threshold bounds without Floor and
  Ceiling, doctor's file split, and more): `docs/superpowers/research/2026-10-07-6d-architecture-reads.md`, "To
  file". Route to the architecture-cleanup task. 2026-10-07 #arch #project:prod-polish
- [ ] #28 xmllint 21504 reports a validity error one line too high once the projected document passes line 65535
  (checked with 65536- and 65540-element stdin documents), so `internal/harness/subsonic/xmlrule/check.go`'s
  line-to-ordinal lookup can name the wrong element on a very large body. Guard the size or correct the offset before
  a large recorded session enters the corpus. Found by slice 2b Task 3's review, 2026-10-07 #harness #project:prod-polish
- [ ] #30 `cmd/harness` cleanup that predates slice 2b, refused at its close as outside the pass. `fetchunidecode.go`
  and `fetchxsd.go` share one cache-fill body. `seeditems.go` and `capture.go` repeat load and capture steps.
  `newCaptureCmd`'s environ and version parameters are a test-only seam. `capture.go:242` and `freeze.go:70` use
  `os.IsNotExist`, which does not unwrap. `capture.go:290` and `synth.go:76` write with `os.WriteFile` where
  `internal/atomicfile` is the house path. Comments in `main.go`, `capture.go`, and `seeditems.go` carry process
  citations. Found by slice 2b's architecture read of `cmd/harness`, 2026-10-10 #harness #arch #project:prod-polish
- [ ] #31 `stack.Start` should verify the oracle image pin and store the result on the `Oracle`: `harness oracle-up`
  never checks the pin today, and `record-oracle` repeats `Start`'s close-on-failure by hand. Found by slice 2b's
  architecture read of `stack`, 2026-10-10 #harness #arch #project:prod-polish
- [ ] #32 `fixture`'s `checkPath` misses a file path that is a directory prefix of another (`a/x.flac` and
  `a/x.flac/y.flac`): `Load` accepts the manifest and `Build` then fails at `MkdirAll`. Refuse it at load. Found by
  slice 2b's architecture read of `fixture`, 2026-10-10 #harness #project:prod-polish
- [ ] #33 For the exported-unused leg's comment-cleanup task: these exports have no caller outside their package
  beyond tests. `grade`'s ledger surface (`State*`, `Class*`, `Pattern`, `ErrInvalidLedger`, `Ledger`, `LoadLedger`,
  `Result`, `Apply`, `Finding`, `Kind*`, `ErrVersionMismatch`) and `stack`'s
  `ErrNoPodman`. Unexport them or name the caller that will use them. Refused at slice 2b's close because the leg is
  staged. `xmlrule`'s five (`ErrVersion`, `ErrSize`, `ErrChecksum`, `ErrNotCached`, `XSDURL`) are unexported (slice 2b Task 12). Found by slice 2b's architecture reads, 2026-10-10 #arch #project:prod-polish
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
- [ ] #55 The first migration after `0005` (`0006`, which the intake spec assigns to `intake-1`) carries R-6d4's
  keep limit: a pre-lock `Inspect` that sees the store busy or under an intent yields no facts, so a migrating
  start prunes to the default. Its tables also join the older-store test helper's table map, or every helper-built
  fixture fails "table already exists". Carried from the store identity pass's routed items, 2026-10-07 #store
  #rung:intake
- [ ] #56 The `id3-tags` reject copy must name a working strip method (`mid3v2 --delete-all` or a tagger), settled
  at the intake copy read. Carried from the alignment pass, 2026-10-11 #upload #rung:intake
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
  declarations. Carried from the alignment pass, 2026-10-11 #docs #project:project-site
- [ ] #63 UNI-L2's failure-alert hook for the unidecode table install. Carried from the alignment pass,
  2026-10-11 #infra #project:packaging
- [ ] #64 The photo checks: `tools/photos/check.test.mjs` fails 2 of 121 (both assume an empty `photos/`), so no gate
  runs it; the photo integrity leg reads `photos/` from disk, not `git ls-files`; and `check.sh`'s header omits that
  leg. Carried from the alignment pass, 2026-10-11 #photos #project:photo-curation
- [ ] #65 Photo licensing follow-ups: a source-URL recheck of the pool, and a blob-retention call for withdrawn
  images. Carried from the alignment pass, 2026-10-11 #photos #project:photo-curation
- [ ] #66 The architecture cleanup: the routed findings in `docs/superpowers/research/*-architecture-reads.md`
  (5d-ii's doctor twelve-flag OR chain first); the ROUTE items in
  `research/2026-09-27-translit-architecture-reads.md` (the four `TestMain` shim copies, the four copies of the
  owner and leaf checks); #24 to #26; the string-deck pass's five stale-comment nits; the roster outcome classes,
  which no independent oracle cross-checks; and the login title argument and W6's h1 fallback, which have no
  rendered proof. The coverage read's 103 coverage gaps ride along, none blocking. Carried from the alignment pass,
  2026-10-11 #arch #project:prod-polish
- [ ] #67 The gate's process-citation leg misses `R-6c` patterns (staged legs 317/127). Carried from the alignment
  pass, 2026-10-11 #gate
- [ ] #68 provision.sh: compute the allowed CIDRs once at top level so errexit applies. Today an all-blank
  `FIREWALL_SSH_CIDRS` falls through to jq's failure, and `hcloud firewall create` may get an empty rules file. A
  behavior change; pairs with #5. Found by the alignment pass, PR #2 #onboarding #project:prod-polish
- [ ] #69 `scripts/check.sh` still says "boundary gate", a retired term. Found by the alignment pass, PR #2 #gate
- [ ] #70 `docs/standing-rules.md` "Gate legs" cites old rung ids and says two legs gate at a comment-cleanup task;
  both legs are warn-only. Found by the alignment pass, PR #2 #gate
- [ ] #71 The rung 7 spec's 7c cites HISTORY line 898 for the 5c-iv handoff; the list is now #57. The first `intake`
  pass repoints it. Found by the alignment pass, PR #2 #rung:intake
- [ ] #72 `label.css` and `label-night.css` set no `--size-field`, though the visual identity spec's "Component
  layer" says every theme block sets it. Found by the alignment pass, PR #2 #web
- [ ] #73 Doc corrections: the `internal/serve/serve.go:66-67` typo ("is answers 404"), and the visual identity
  spec's "Copy" section, which still carries the old note about W's section 19. Found by the alignment pass, PR #2
  #project:prod-polish
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
  2026-10-12 #project:prod-polish
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

## Done

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
  #infra #project:prod-polish

- [x] #11 Resolve the docker compose leg's `command -v docker` skip: retired with the leg
  itself (`d1e93e9` removed the compose leg from `scripts/check.sh`), 2026-10-05 #infra
- [x] #29 The session credential scan treated every string leaf holding `?` as a URL query: it now splits the leaf
  on `?`, `&` and `;`, reads only `key=value` fields, and refuses nothing for prose (`be9abf5`), 2026-10-10 #harness
- [x] #22 `harness capture` overwrote an existing case with no existence check: capture and synth now share one
  overwrite guard (`e561c32`, slice 2b Task 1), 2026-10-10 #harness
- [x] #27 `check.sh --scope-changed` ran the rendered leg for a harness-only change: the conform and rendered scopes
  are derived apart (`e561c32`, slice 2b Task 1), 2026-10-10 #gate
