# Draft HISTORY entry: rung 5, pass 5b-ii (auth handlers, mail seam, limiters, close)

**Draft only. The conductor folds this into `docs/HISTORY.md` at the merge, path-limited.**

---

## Rung 5, pass 5b-ii (the roster, the mail seam, the rate limiters, the origin check, the auth handlers, the security fold): closed 2026-09-23, lane 1

Executed against `docs/plans/2026-09-12-rung-5b-ii-auth-handlers.md`
(revision 2.1, launch-ruled 2026-09-23) in the main tree, from `1d4ed29`
(5b-i's merge). Task 0 (`opus`, fresh context) ran at `f75b0d9`
(`research/handoffs/2026-09-12-5b-ii-task0-staleness.md`); the Go tree
was unchanged at `dfda0ff`, so its evidence stands unrefreshed through
the close. Its one live finding, that every 5b-i `HISTORY:` pointer in
this plan sits 47 lines low, is applied below (launch ruling c) rather
than re-derived.

Tasks 1 through 11 ran across four segments: A (Tasks 1-3, closed
`aae338c`), B (Tasks 4-6, closed `dfa7155`/`3ca0550`), C (Tasks 7-8,
closed `c09bf44`, simplifier `105ad50` + `95e7607`), and D (Tasks 9-11,
closed `c2613ba`, simplifier `08befea`), followed by the pass-end
`web-auth-security-reviewer` (`46e37d1`) and this close (Task 12).

### The gate evidence

`go test ./...` at the merged tip: **2m23.8s** wall (`internal/harness/conform`
at 140.9s and `internal/harness` at 11.6s dominate; every other package
under 2.2s). Package-scoped: `internal/auth` **2.1s**, `internal/identity`
**0.006s**, `internal/notify` **0.027s**, `internal/serve` **1.1s**
(all `-count=1`). `bash scripts/check.sh` ran green (exit 0) through
`cairn-run-gate` at this close; the wrapper printed no wall-clock total
at the time, so this draft's first round left the claim as a bare
"green" (the first thing 5b-i's reviewer escalated, and this round's
own escalation). **Corrected**: re-running the same gate at this close's
own commit with `time CAIRN_GATE_LANE=light cairn-run-gate 'bash
scripts/check.sh'` measured **1m10.099s real** (`6m35.341s` user,
`0m3.586s` sys) for a clean pass. This is a fresh measurement at this
close's own commit, not a reconstruction of the historical merge's own
wall time, which no artifact recorded.

`.gitleaksignore` and `env/` are byte-identical across `1d4ed29..HEAD`
(`git diff` empty for both). `go.mod` gained exactly the two dependencies
R26 budgeted: `github.com/wneessen/go-mail v0.8.1` (MIT, Task 6) and
`golang.org/x/time v0.16.0` (Task 7).

Eleven existing test files changed, none weakened:

| Existing test file | Why it changed |
| --- | --- |
| `internal/identity/identity_test.go` | `NormalizeAddress` gained the B1/M1-adjacent control-byte and space refusal (Task 1) |
| `internal/auth/nonce_test.go` | B1's retention-vs-expiry fix and Task 2's single-unreadable-record fix both land here |
| `internal/auth/epoch_test.go` | M2's backward-bump lock and the subscriber-panic recovery (Tasks 1-2) |
| `internal/auth/token_test.go` | M3/M5's strict-decode, CR/LF, and bounded-allocation fixes; additions only, no deletions |
| `internal/auth/session_test.go` | m1/m2's Max-Age-plus-clamp fix (Task 2) |
| `internal/auth/store_test.go` | Task 2's 0700 refusal, dropped `Store.Write`, and `e713a09`'s correction of which entry point `NonceStore`/`EpochStore` write through |
| `cmd/dubplate-server/doctor_test.go` | Task 2's flattened row-status fix and Task 4's roster row |
| `cmd/dubplate-server/auth_test.go` | Task 3's six remaining thresholds and the `authConfig` struct replacing four positional durations |
| `cmd/dubplate-server/check_test.go` | the check verb's fixture environment gained a roster so config load resolves |
| `cmd/dubplate-server/serve_test.go` | Task 11 wires the auth lifecycle into `newServeShell` |
| `internal/serve/serve_test.go` | Task 8's origin check joins the composition chain `Handler()` builds |

### What landed

Task 1 (`f90c9ea`, `opus`) folded the 5b-i blocker and four reachable
majors: nonce records now live for the later of retention and the
token's own expiry (B1), `NormalizeAddress` is a fixed point that
refuses control bytes and spaces (M1), `Bump` never moves an epoch
backward and holds one lock across its read and write (M2), and
`splitSigned` decodes strictly, refuses CR and LF, and bounds its split
(M3/M5). This satisfies R-5b17: the fold ran as Tasks 1-2 of this pass
rather than a separate remediation pass, so every later task built
against corrected code.

Task 2 (`772eb22`, then `e713a09` fix, then `aae338c` second fix) folded
the 5b-i minors and fixed `doctor`'s rendering: session cookies carry
`Max-Age` alongside `Expires`, `RefreshSession` clamps to anchor-plus-cap,
a cookie-clearing helper matches the `__Host-` cookie's attributes, the
leaf refuses a wider-than-0700 existing directory, `Store.Write` is
gone, `RecordID` mixes in a per-kind label, an unreadable nonce record
no longer blocks startup, and epoch subscriber notifications run under
recover and a bounded wait. The magic-link thresholds gained floors and
ceilings, and the public-URL check refuses userinfo, a query, and a
fragment. `identitySourceStatus` reruns the selected source's `Validate`
hook so a relational refusal (R31) surfaces on the source's own row.
The reviewer's blocking finding (a `strings.Contains` assertion that a
line-count regression could still pass) was fixed with an exact-count
test reusing the two-secrets-below-floor fixture, accepted on the second
round (conductor-ruled: the narrow fix, not a broader doctor rewrite).

Task 3 (`323a817`, then `e0a3aa8` fix) registered the six remaining
auth thresholds: the four rates, the body cap, and the unknown-address
answer, each validated against a positive floor and a stated ceiling.
`DUBPLATE_SMTP_FROM` now parses as one RFC 5322 address with no CR/LF
at load. `resolveAuthConfig` returns an `authConfig` struct in place of
four positional durations (launch ruling b), since a sixth and seventh
threshold made the positional list unreadable. This lands R-5b22 (the
`KindString` threshold) and R-5b23 (the map bound and sweep period as
package constants, not thresholds). The fix round traced every one of
ten declared defaults to its ruled source in a doc comment and ran the
table against `productionRegistry()` (launch ruling d).

Task 4 (`d9d422e`, then `dfa7155` fix) added `internal/auth/roster.go`:
the roster fails closed on load (R2), validates B9's field list, and
exposes one `normalizeMemberAddress` resolution point (R-5b19's JSON
envelope with a schema version). `OpenRoster` is the read-open;
`CreateRoster` is the write-open the CLI bootstrap will call (R-5b20),
creating the file at 0600 and refusing an existing one, but creating no
directory (R-5b20's own deferral of the pre-create to 5d).
`config/roster.example.json` exists at this commit. The fix round
stopped `Lookup` from wrapping `net/mail`'s parse error (which echoed
an address) and switched the control-character check to
`unicode.IsControl` so it also refuses C1 controls.

Task 5 (`cb10873`, then `dfa7155` fix) added the write path: `AddMember`
validates against B9, takes the store's lock, re-checks the normalized
address and `streamer_username` against the roster as loaded, writes
atomically through `atomicfile.WriteStream`, re-reads the written file,
and logs one structured `slog` audit record (R-5b21) naming the actor,
channel, address, admin flag, and client IP. The admin flag is refused
on every channel, a service token included, by its own sentinel
(`internal/auth/roster.go:323-331`) — the conductor's ruling that the
invariant lives in the write path, not in 5c's caller; 5c's CLI
`member add --admin` brings the admin grant together with its own
caller. The fix round (`dfa7155`) removed the write path's original
service-token exception, deleting
`TestAddMember_AdminFlagAllowedOnServiceTokenChannel` and folding a
service-token admin row into `TestAddMember_ByteIdentityOnFailure`
(`internal/auth/roster_test.go:445-449`), and gave `doctor` a row that
reports a missing roster as a failure when a selected source requires
one.

Task 6 (`78779eb`, then `3ca0550` fix) added `internal/notify`, adopted
from `_salvage/internal/notify` on merit under fresh review and
restructured so subject and body arrive as data, with
`github.com/wneessen/go-mail` v0.8.1 (MIT) as the SMTP transport. The
header-injection guard covers the recipient, subject, and composed
`From` before any message is built. The fix round dropped the
recipient and `From` address from every error string (go-mail's own
parse/send errors quote the address, which Global Constraints forbid
outside the roster audit record) and added `ErrInvalidFrom`,
`ErrInvalidRecipient`, and `ErrSend` sentinels for `errors.Is` testing.

Task 7 (`496f8c1`) added the three rate limiters: link requests per
address, link requests per client, and consumption attempts per client,
each over its own LRU-bounded keyed map of `golang.org/x/time/rate`
buckets (R-5b24, three key spaces rather than one shared map), plus a
global outbound link-mail ceiling. Adopts `golang.org/x/time` v0.16.0.

Task 8 (`c09bf44`) added the origin check: `internal/serve/origin.go`
wraps `net/http.CrossOriginProtection` so a state-changing request
carrying neither `Sec-Fetch-Site` nor `Origin` is refused, inserted in
`Shell.Handler` between normalization and prefix routing (R-5b27) so
every registered subtree and the fallthrough sit behind it. The wrapper
keeps its own bypass mux beside the stdlib value's, because the stdlib
admits the neither-header case before consulting its own bypass
patterns.

The segment-C simplifier (`105ad50`, then `95e7607`) made
`validateMemberFields` a method on `memberRecord` (its own package's
type, no other caller) and aligned `doc.go`'s and `Shell.Handler`'s
composition-order comments on the same five steps in the same order,
dropping a process citation in the second pass.

Task 9 (`0cee9e8`, then `d679ecc` fix) added the verify-identity seam:
`MagicLinkSeam` verifies a session cookie's MAC and expiry, re-checks
the address against the roster and the revocation epoch on every call,
and refreshes the session only for a same-origin request. The admin
flag travels only from the roster record, never the cookie. The fix
round moved the permit/refuse contract sentence onto the method it
actually describes (it did not hold for `Read`, where a nil error means
the read succeeded, not that a request was permitted) and stated the
add-member carve-out on `VerifiedIdentity.Admin` explicitly.

Task 10 (`644dfba`) added the request-link, consume-link, and sign-out
handlers, accepted first time. All twelve named mutations fired
(the plan's own count for this task); test-first was not kept for this
task, a decision recorded rather than absorbed silently.

Task 11 (`c2613ba`) wired `newServeShell` to open the auth store, the
nonce and epoch stores, and the roster, build the rate limiter, the
mailer, and the magic-link seam, construct the handlers, and mount
request-link, consume-link, and sign-out under the owned prefix behind
a POST-only wrapper, all before `serve.Run` opens a listener. Accepted
first time. `internal/auth/nonce.go` gained `SweepInterval` and
`RunSweep`, driven by the serve verb's own goroutine bound to the
verb's context (R-5b14's interval driver, deferred by 5b-i to this
pass).

The segment-D simplifier (`08befea`) hoisted the generic-answer send's
`context.WithoutCancel(r.Context())` call out of the goroutine closure
into the caller, behavior-neutral.

### The contract coverage, D1 through D13

Each row names the clause's Go site (file and symbol) and the test that
proves it. No D-clause in this pass has a stated no-test reason; every
one has a running test. D5 and D11 carry a note on the limits of what
their tests currently reach.

| Clause | Go site | Proving test |
| --- | --- | --- |
| D1 (a consumed nonce's record outlives its token) | `internal/auth/nonce.go:138` `NonceStore.Insert` | `TestRecordOutlivesItsToken` (`internal/auth/nonce_test.go`) |
| D2 (address normalization is a fixed point) | `internal/identity/identity.go:42` `NormalizeAddress` | `TestNormalizeAddressIsAFixedPoint` (`internal/identity/identity_test.go`); `TestMintLinkRefusesAnUnrevocableAddress` (`internal/auth/token_test.go`) |
| D3 (the revocation epoch moves forward only) | `internal/auth/epoch.go:103` `EpochStore.Bump` | `TestBackwardBumpKeepsTheRevocation`, `TestConcurrentBumpsKeepTheLater` (`internal/auth/epoch_test.go`) |
| D4 (one credential has exactly one verifying string) | `internal/auth/token.go:259` `splitSigned` | `TestLinkHasOneEncoding`, `TestSeparatorFloodAllocatesLittle` (`internal/auth/token_test.go`) |
| D5 (`doctor` never prints all-green under a `Load` refusal) | `cmd/dubplate-server/doctor.go:156` `rowStatuses`, `:268` `identitySourceStatus` | `TestDoctorCmd_NoRunReportsAllGreenOnLoadError`, `TestRowStatuses_FlattensNestedJoin` (`cmd/dubplate-server/doctor_test.go`). **Reached only for the load-bearing property**: the security review's finding 11 shows no test yet proves the row *names the offending key* for a bad identity-source value, which M4's partial status and gap-list item are the stated reason for (routed to 5c). |
| D6 (the roster fails closed, validates B9, one resolution point) | `internal/auth/roster.go:129` `OpenRoster`, `:147` `CreateRoster` | `TestOpenRoster_LoadRefusals`, `TestOpenRoster_ValidationRows`, `TestRoster_TwoEntryPoints`, `TestNormalizeMemberAddress_SingleResolutionPoint` (`internal/auth/roster_test.go`) |
| D7 (a roster write is all-or-nothing, serialized, visible without a restart) | `internal/auth/roster.go:328` `Roster.AddMember` | `TestAddMember_ByteIdentityOnFailure`, `TestAddMember_ConcurrentDistinctMembersBothSurvive`, `TestAddMember_ConcurrentStreamerUsernameCollision`, `TestAddMember_WriteReReadAndMode`, `TestAddMember_AuditRecord` (`internal/auth/roster_test.go`) |
| D8 (mail is a general seam with a header-injection guard) | `internal/notify/notify.go:89` `Mailer.Send` | `TestSend_HeaderInjectionGuard` (`internal/notify/notify_test.go`) |
| D9 (no address or client drives unbounded mail; the map is bounded) | `internal/auth/limit.go` `keyedLimiter`, `Limiter.AllowLinkRequest`/`AllowConsumption`/`AllowLinkMail` | `TestKeyedLimiterBoundHoldsUnderFlood`, `TestKeySpacesDoNotShareBuckets` (`internal/auth/limit_test.go`) |
| D10 (one origin rule covers every cookie-authenticated surface) | `internal/serve/origin.go:47` `originCheck`, `:71` `newOriginCheck` | `TestOriginCheck`, `TestOriginCheckBearerBypass` (`internal/serve/origin_test.go`) |
| D11 (a verified identity carries its channel; admin is roster-sourced) | `internal/auth/seam.go:105` `MagicLinkSeam.VerifyIdentity` | `TestMagicLinkSeam_VerifyIdentity`, `TestMagicLinkSeam_VerifyIdentity_EpochRevokedAfterRefresh`, `TestMagicLinkSeam_VerifyIdentity_CrossOriginDoesNotRefresh` (`internal/auth/seam_test.go`). **Tests exist and pass**; what is missing is not test coverage but the mutation-run record for Task 9's four named mutations, lost when the fix round superseded the round-1 report (see the mutation table's Task 9 row below). |
| D12 (the consumption order is fixed; no session mints before `Insert`) | `internal/auth/handlers.go:260` `handlers.consumeLink` | `TestConsumeLink_Refusals`, `TestConsumeLink_LimiterPrecedesMAC`, `TestOversizeBodyRefusedBeforeParse`, `TestLogsCarryNoCredential` (`internal/auth/handlers_test.go`) |
| D13 (the sweep runs on a schedule the serve verb owns) | `internal/auth/nonce.go:102` `SweepInterval`, `:112` `RunSweep` | `TestRunSweepRunsOnTickWithoutWaitingSweepInterval` (`internal/auth/nonce_test.go`); `TestStartNonceSweep_RunsAndStopsWithContext`, `TestNewServeShell_SweepDriverStopsWithTheGivenContext` (`cmd/dubplate-server/serve_test.go`) |

### The mutation table across Tasks 1 through 11

Built from the five pass-execute workflow records
(`wf_e12d88dd-d69.json` Tasks 1-3, `wf_fe30c4db-249.json` Tasks 4-6,
`wf_f6babece-2e7.json` Tasks 7-8, `wf_cd9bc959-990.json` Tasks 9-10,
`wf_7f3f9b8d-26c.json` Task 11), reading each task's `mutationLedger`
from its accepted (or final-fix-round) implementer report. Every row's
count matches its list.

**Task 1 — 8 mutations, 8 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Write `now.Add(retention)` rather than the max | `internal/auth/nonce.go` `Insert` | yes | `TestRecordOutlivesItsToken` |
| 2 | Drop the re-parse from the normalizer | `internal/identity/identity.go` `NormalizeAddress` | yes | `TestNormalizeAddress`/`TestNormalizeAddressIsAFixedPoint` quoted-comma subtests (2); `TestMintLinkRefusesAnUnrevocableAddress` (1) |
| 3 | Drop the control-character filter | `internal/identity/identity.go` `NormalizeAddress` | yes | `TestNormalizeAddress` TAB and space subtests (2) |
| 4 | Write `now` unconditionally in `Bump` (drop the backward/equal guards) | `internal/auth/epoch.go` `advance` | yes | `TestBackwardBumpKeepsTheRevocation`, `TestConcurrentBumpsKeepTheLater` |
| 5 | Move `Bump`'s read outside the mutex | `internal/auth/epoch.go` `advance` | yes | `TestConcurrentBumpsKeepTheLater` (fails with and without `-race`, 30/30 runs; clean row passed 50/50 under `-race`) |
| 6 | Decode with the non-strict decoder | `internal/auth/token.go` `fieldEncoding` | yes | `TestLinkHasOneEncoding`, `TestSessionHasOneEncoding` (3 substitutions each) |
| 7 | Restore the unbounded `strings.Split` | `internal/auth/token.go` `splitSigned` | yes | `TestSeparatorFloodAllocatesLittle` (~16.8 MB against a 1 MiB ceiling) |
| 8 (extra, not named in the plan) | Drop the CR/LF refusal | `internal/auth/token.go` `splitSigned` | yes | `TestLinkHasOneEncoding`, `TestSessionHasOneEncoding` (4 CR/LF variants each) |

**Task 2 — 3 mutations, 2 fired, 1 unfired (reason recorded).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | The fix round's own revert: drop the `Validate` rerun from `identitySourceStatus` | `cmd/dubplate-server/doctor.go` `identitySourceStatus` | yes | `TestDoctorCmd_NoRunReportsAllGreenOnLoadError`'s magic-link-row assertion fails |
| 2 (named 1) | Unwrap one level instead of recursing | `cmd/dubplate-server/doctor.go` `flattenErrors` | yes | `TestRowStatuses_FlattensNestedJoin` (`B_KEY` row differs) |
| 3 (named 2) | Drop the `Optional`-consultation loop, leaving the `Validate` rerun in place | `cmd/dubplate-server/doctor.go` `identitySourceStatus` | **no** | `TestDoctorCmd_ShortPreviousSecretMakesSourceUnsatisfied` still passes: the `Validate` rerun this dispatch added independently checks the one `Optional` key magic-link currently declares, so no test can currently distinguish the `Optional` branch from the `Validate` branch. The load-bearing property (no all-green row under a `Load` refusal) still holds through the other path. A future `Optional` key that `Validate` does not itself check would restore this mutation's bite. |

**Task 3 — 2 mutations, 2 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Drop the `authRequestBodyCap` floor check | `cmd/dubplate-server/auth.go:313` `checkIntBounds` | yes | `TestValidateMagicLinkSource_RateAndCapBounds` (4 floor subtests) |
| 2 | Widen the unknown-address accepted set to also accept `"silent"` | `cmd/dubplate-server/auth.go:279` `validateMagicLinkSource` | yes | `TestValidateMagicLinkSource_UnknownAddressAnswer`/`unrecognized` |

**Task 4 — 8 mutations, 8 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Return an empty roster and nil error for a missing file | `internal/auth/roster.go` `OpenRoster` | yes | `TestOpenRoster_LoadRefusals`/`missing_file` |
| 2 | Compare addresses without normalizing | `internal/auth/roster.go` `validateRoster` | yes | `TestOpenRoster_ValidationRows`/`a_duplicate_address_differing_only_in_case` |
| 3 | Drop the `streamer_username` uniqueness check | `internal/auth/roster.go` `validateRoster` | yes | `TestOpenRoster_ValidationRows`/`a_duplicate_streamer_username` |
| 4 | Compare the length bound against rune count rather than byte length | `internal/auth/roster.go` `validateRoster` | yes | `TestOpenRoster_ValidationRows`/`display_name_one_byte_over_the_length_cap` |
| 5 | Add a second address-to-identity derivation | `internal/auth/roster.go` `Lookup` | yes | `TestNormalizeMemberAddress_SingleResolutionPoint` |
| 6 | Let the write-open succeed against an existing file | `internal/auth/roster.go` `CreateRoster` | yes | `TestRoster_TwoEntryPoints` |
| 7 | Let the read-open tolerate an absent file once the write-open exists | no distinct site (collapses onto mutation 1's edit) | yes | Same evidence as mutation 1; reported as a finding, since the design has no separate code path this mutation could target |
| 8 | Drop the `doctor` roster row | `cmd/dubplate-server/doctor.go` `RunE` | yes | `TestDoctorCmd_RosterRow`/`roster_with_a_duplicate_streamer_username` |

**Task 5 — 7 mutations, 6 fired, 1 unfired (reason recorded).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | Write before validating | `AddMember` (dropped early returns) | yes | `TestAddMember_ByteIdentityOnFailure` (3 of 5 subtests) |
| 2 | Take the lock after the re-check | `AddMember` (moved `r.mu.Lock()`) | yes | `TestAddMember_ConcurrentStreamerUsernameCollision`, `-race -count=20` |
| 3 | Use `atomicfile.WriteFile` rather than `WriteStream` | `AddMember` | **no** | `go test -race ./internal/auth/...` stayed green; per the plan this is a review-caught defect (the durability doc comment stops matching the call), not a test-caught one, and no test was written to force it |
| 4 | Skip the re-read | `AddMember` (dropped the post-write reparse) | yes | `TestAddMember_WriteReReadAndMode`, `TestAddMember_ConcurrentDistinctMembersBothSurvive`, `TestAddMember_ConcurrentStreamerUsernameCollision` |
| 5 | Drop the channel from the audit record | `AddMember` (logger call) | yes | `TestAddMember_AuditRecord` |
| 6 | Accept the admin flag on any assertion channel | `AddMember` (guard forced `if false`) | yes | `TestAddMember_ByteIdentityOnFailure` (cookie-, header-, and service-token-channel admin rows) |
| 7 | Let the read-open tolerate an absent file | `OpenRoster` | yes | `TestAddMember_WriteReReadAndMode` |

**Task 6 — 4 mutations, 4 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Guard the recipient and not the subject | `internal/notify/notify.go` `Send` | yes | `TestSend_HeaderInjectionGuard`/subject CR and LF |
| 2 | Guard only LF, not CR | `internal/notify/notify.go` `guardHeaderValue` | yes | `TestSend_HeaderInjectionGuard`/recipient CR, subject CR |
| 3 | Guard after composing the message | `internal/notify/notify.go` `Send` (guard-vs-compose order) | yes | `TestSend_HeaderInjectionGuard`/recipient CR and LF |
| 4 | Route port 465 through STARTTLS | `internal/notify/notify.go` `Send` (TLS branch selection) | yes | `TestSend_ImplicitTLS465` |

**Task 7 — 6 mutations, 5 fired, 1 unfired (deferred to Task 10's own ledger).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| M1 | Remove the map's bound | `internal/auth/limit.go` `keyedLimiter.allow` | yes | `TestKeyedLimiterBoundHoldsUnderFlood`, `TestKeyedLimiterEvictsLeastRecentlyUsed`, `TestKeyedLimiterConcurrent` |
| M2 | Share one `keyedLimiter` across all three limiters | `internal/auth/limit.go` `NewLimiter` | yes | `TestKeySpacesDoNotShareBuckets` (both subtests); 3 more test functions, 5 subtests total |
| M3 | Evict most-recently-used instead of least | `internal/auth/limit.go` `keyedLimiter.allow` | yes | `TestKeyedLimiterEvictsLeastRecentlyUsed`, `TestKeyedLimiterConcurrent` |
| M4 | Key the per-client limiter on the wrong source | call site in Task 10 | **no (here)** | No test in this package can reach it: `Limiter` takes the client address as a string and derives none, so this is a call-site defect. The plan routes this mutation to Task 10, where it is proved (see Task 10 row 12). |
| M5a | Only the address key gates | `internal/auth/limit.go` `AllowLinkRequest` | yes | `TestLinkRequestKeysTripIndependently`/`many_addresses_from_one_client`, `TestKeySpacesDoNotShareBuckets`/`link requests per client` |
| M5b | Only the client key gates | `internal/auth/limit.go` `AllowLinkRequest` | yes | `TestLinkRequestKeysTripIndependently`/`one_address_from_many_clients`, `TestLinkRequestNormalizesAddress` |

**Task 8 — 7 mutations, 6 fired, 1 unfired (reviewer-graded, no test can reach).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | Drop the tightening; use the stdlib default | `internal/serve/origin.go` `allowed()` | yes | `TestOriginCheck`, `TestOriginCheckBearerBypass`, `TestOriginCheckCoversSubtreeAndFallthrough`, `TestOriginCheckRunsAfterNormalize`, `TestOriginRefusalEmitsTheRefusalRecord` (5 top-level, 6 subtests) |
| 2 | Move the check inside the prefix's routed handler | `internal/serve/serve.go` `Handler()` | yes | `TestOriginCheckCoversSubtreeAndFallthrough`, `TestOriginRefusalEmitsTheRefusalRecord` |
| 3 | Move the check outside `Normalize` | `internal/serve/serve.go` `Handler()` | yes | `TestOriginCheckRunsAfterNormalize`, `TestEchoMethodsRefusedBeforeRouting` (4 rows) |
| 4 | Let a bypassed path read the session cookie | contract on `originCheck`'s doc comment | **no** | No test can reach it: no bearer-authenticated path exists yet at this rung. Graded by a reviewer against the doc comment, per the plan's own instruction. |
| 5 | Consult the bypass set after the neither-header denial | `internal/serve/origin.go` `allowed()` | yes | `TestOriginCheckBearerBypass`/`bypassed path, POST carrying neither header, is allowed` |
| 6 | Register through `AddInsecureBypassPattern` alone | `internal/serve/origin.go` `addBearerBypass()`/`allowed()` | yes | `TestOriginCheckBearerBypass`/same subtest |
| 7 (extra, not named in the plan) | Drop the `AddInsecureBypassPattern` call | `internal/serve/origin.go` `addBearerBypass()` | yes | `TestOriginCheckBearerBypass`/`bypassed path, POST carrying a cross-origin Origin, is allowed` |

**Task 9 — mutation count and outcomes not recoverable from the durable record.**

The plan names 4 mutations for Task 9 (`plan:2644-2648`): reading the
admin flag from the session value, skipping the per-request roster
re-check, treating a non-sentinel error as "not refused", and
refreshing unconditionally. The round-1 implementer report (commit
`0cee9e8`) is the report that would carry this ledger, but the task's
one fix round (`d679ecc`, doc comments only) produced the accepted
report the workflow record retains, and that report carries no
`mutationLedger` field (confirmed: the `diff-reviewer`'s own
`review:9` non-blocking note independently states "this round's report
has no mutation ledger"). The round-1 report's full JSON was not
retained beyond a ~400-byte truncated preview in
`wf_cd9bc959-990.json`, which cuts off before any mutation content.
**This is a gap in the durable record, not a claim that the mutations
were skipped**: `internal/auth/seam_test.go` (`TestMagicLinkSeam_VerifyIdentity`
and its three siblings) exercises the same four properties the named
mutations target, and Task 9 was accepted on its second round. No
fired/unfired count is stated here because none can be reconstructed
from what survives.

**Task 10 — 12 mutations, 12 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Insert the nonce before verifying the MAC | `internal/auth/handlers.go` `consume` | yes | `TestConsumeLink_Refusals` (5 rows) |
| 2 | Mint the session before `Insert` returns nil | `internal/auth/handlers.go` `consume`/`consumeLink` | yes | `TestConsumeLink_Refusals`/`replayed_link`, `TestConsumeLink_SignsIn`, `TestLogsCarryNoCredential` |
| 3 | Resolve the roster after `Insert` | `internal/auth/handlers.go` `consume` | yes | `TestConsumeLink_Refusals`/`address_not_on_the_roster` |
| 4 | Branch on sentinels alone (`CheckLink`) | `internal/auth/handlers.go` `consume` | yes | `TestConsumeLink_Refusals`/`epoch_unreadable` |
| 5 | Parse before capping the body | `internal/auth/handlers.go` `readForm` | yes | `TestOversizeBodyRefusedBeforeParse` (3 rows) |
| 6 | Return a distinguishable outcome per failure cause | `internal/auth/handlers.go` `consumeLink` | yes | `TestConsumeLink_Refusals` byte-identity check |
| 7 | Key the per-client limiter on the submitted string (Task 7's M4, landed here) | `internal/auth/handlers.go` `requestLink` | yes | `TestRequestLink_ClientFloodSendsNoFurtherMail`, `TestRequestLink_ClientAddressIsTheShells`/`untrusted` |
| 8 | Log `RawQuery` on the consumption route | `internal/auth/handlers.go` `consumeLink` logger | yes | `TestLogsCarryNoCredential` |
| 9 | Move the consumption limiter after the MAC check | `internal/auth/handlers.go` `consumeLink` | yes | `TestConsumeLink_LimiterPrecedesMAC` |
| 10 | Surface the send failure under `generic` (synchronous, 500 on failure) | `internal/auth/handlers.go` `requestLink` generic branch | yes | `TestRequestLink_GenericAnswersAlike`/`send_fails`, `TestRequestLink_GenericDoesNotWaitOnSend` |
| 11 | Answer under `generic` only after the send completes | `internal/auth/handlers.go` `requestLink` generic branch | yes | `TestRequestLink_GenericDoesNotWaitOnSend` (5.00s deadline) |
| 12a | Key on the forwarding header where no proxy is trusted | `internal/auth/handlers.go` `requestLink` | yes | `TestRequestLink_ClientAddressIsTheShells`/`untrusted_peer` |
| 12b | Key on the peer address where a trusted proxy is configured | `internal/auth/handlers.go` `requestLink` | yes | `TestRequestLink_ClientAddressIsTheShells`/`trusted_proxy` |

(12a/12b are one named mutation with two branches in the ledger; the
plan's own count of twelve for this task counts them as one, matching
"all twelve named mutations fired" above.)

**Task 11 — 4 mutations, 3 fired, 1 structurally unreachable (reason recorded).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | Mount the handlers outside the owned prefix | `cmd/dubplate-server/serve.go` `newServeShell` | yes | `TestNewServeShell_MountsTheThreeAuthRoutes`/`consume-link` |
| 2 | Start the sweep driver without the verb's context | `cmd/dubplate-server/serve.go` `newServeShell` | yes | `TestNewServeShell_SweepDriverStopsWithTheGivenContext` (added specifically because the pre-existing `TestStartNonceSweep_RunsAndStopsWithContext` calls `startNonceSweep` directly and did not catch this on the first pass) |
| 3 | Open the stores after listening | `cmd/dubplate-server/serve.go` `newServeShell` (`auth.OpenRoster` err check) | yes | `TestNewServeShell_MissingRosterRefusesBeforeConstruction` |
| 4 | Register a bypass pattern for the fallthrough | n/a — no site in this task's Files | **no** | Structurally unreachable: `internal/serve.Shell` exposes no exported method to register a bearer bypass (`originCheck.addBearerBypass` is unexported on an unexported type), and this task's Files (`cmd/dubplate-server/serve.go`, `internal/auth/nonce.go`) hold no code path that could register one; `internal/serve/origin.go` is outside this task's Files. `TestNewServeShell_StateChangingFallthroughRequestIsDenied` proves the absence behaviorally, but the mutation cannot be introduced from this task's own surface. |

### The twelve resolutions (R-5b17 through R-5b28)

**The plan's grep-recoverable-marker mechanism did not hold for this
pass.** `plan:3372-3378` reads this table off `grep -rn 'ruled,
conductor 2026-09-12, R-5b'`. That grep, and the looser
`grep -rn 'R-5b1[7-9]\|R-5b2[0-8]'`, together find markers for **3 of
the 12** resolutions: R-5b19 (`internal/auth/roster.go:19`, `:98`),
R-5b20 (`internal/auth/roster.go:140`), and R-5b22
(`cmd/dubplate-server/auth.go:161`, `cmd/dubplate-server/auth_test.go:460`).
**The other 9 (R-5b17, R-5b18, R-5b21, R-5b23 through R-5b28) carry no
marker anywhere in `cmd/` or `internal/`.** The table below states
"no marker" on each of those nine rather than naming a site that does
not carry one; the "Landed" column still names the file where the
resolution's effect lives, since that is verifiable independent of any
marker.

| Resolution | Decided | Marker | Landed (effect) | Confirmed by execution |
| --- | --- | --- | --- | --- |
| R-5b17 | 5b-i's security fold is Tasks 1-2, not a separate pass | no marker | Tasks 1-2 | Confirmed: B1/M1/M2/M3/M5 fixed before Task 3 dispatched |
| R-5b18 | 5b-i's security read is discharged; this pass's own read covers both halves at close | no marker | pass-end review, `46e37d1` | Confirmed: the review's scope line covers `1d4ed29..HEAD`, both halves |
| R-5b19 | roster file is `{"schema_version": 1, "members": [...]}`, not a bare array | `internal/auth/roster.go:19`, `:98` | same file | Landed as decided |
| R-5b20 | the store creates the file, never the directory; pre-create is 5d's | `internal/auth/roster.go:140` (`CreateRoster`) | same file | Landed as decided; `config/roster.example.json` created here rather than in 5d (recorded, spec correction below) |
| R-5b21 | the audit record is a structured `slog` record; no audit file opens | no marker | `internal/auth/roster.go` (`AddMember`, Task 5) | Landed as decided |
| R-5b22 | the unknown-address answer is `KindString`, default `kind`, values `kind`/`generic` | `cmd/dubplate-server/auth.go:161`, `auth_test.go:460` | `cmd/dubplate-server/auth.go:197-280` | Landed as decided |
| R-5b23 | the limiter map bound and sweep period are package constants, not thresholds | no marker | `internal/auth/limit.go`, `internal/auth/nonce.go` | Landed as decided |
| R-5b24 | three limiters, three key spaces, not one shared map | no marker | `internal/auth/limit.go` (`keyedLimiter`, three instances) | Landed as decided; **the security review's finding 3 shows the per-address key space is still charged before the per-client check refuses, an ordering gap R-5b24 did not itself cover** |
| R-5b25 | handlers are `http.Handler` values from one dependency struct, no package-level route registration | no marker | `internal/auth/handlers.go` | Landed as decided |
| R-5b26 | no redirect-target validator; R33 travels as a constraint to rung 12 | no marker | disposed, no function added | Confirmed: Task 0 item 10 and this close both grep `internal/auth` and find none; the `b91bae1` erratum's contrary claim is corrected below |
| R-5b27 | the origin check inserts between normalization and prefix routing | no marker | `internal/serve/serve.go:159-172` (`Handler`) | Confirmed by reading `Handler`'s wrap order: strip, Normalize, refuseEchoMethods, origin check, routed |
| R-5b28 | the generic answer's ordering: `kind` governs R9, `generic` answers before the send | no marker | `internal/auth/handlers.go` (Task 10) | Confirmed by the security review: "the generic answer is identical in bytes and in measured timing" |

### The seven conductor-owed folds

Ruled at dispatch (plan `:22-56`); execution's own evidence for each:

1. **The ceiling.** Ruled 7.5M, one pass, rather than cutting at
   dispatch. Confirmed: the pass closed at about 6.4M, under the
   ceiling with no split.
2. **R3's `doctor` roster collision row lands in Task 4** as its fifth
   deliverable. Confirmed: `d9d422e`/`dfa7155` add the roster row to
   `doctor.go` inside Task 4's own commits, not routed as a declared
   gap.
3. **R9 against fork 4's `generic` mode (R-5b28): the narrow reading**,
   R9 governs `kind` and `generic` answers before the send. Landed at
   Task 10; the security review measured identical bytes and timing for
   the generic answer.
4. **The admin-flag invariant enforces in the store's write path**, not
   in 5c's caller. Confirmed: `AddMember`
   (`internal/auth/roster.go:323-331`, Task 5, fixed at `dfa7155` to
   drop the write path's original service-token exception) refuses
   `Admin` on every channel, a service token included, and the security
   review's "attacked and found sound" list confirms the seam reads
   `Admin` only from the roster, never the request.
5. **Task 8's model is `opus` unconditionally.** Confirmed: `c09bf44`'s
   trailer credits Claude Opus 5.5.
6. **M5's failing-first test ships the named fallback if the allocation
   measurement cannot be made to fail against the unfixed code.**
   The fallback was not needed: `TestSeparatorFloodAllocatesLittle`
   (`internal/auth/token_test.go`) measures `AllocedBytesPerOp()`
   against a ceiling derived from the flood's own length, which the
   unbounded pre-fix split would exceed, so the fallback this ruling
   authorized never fired. Recorded here rather than as a declared gap
   (see the corrected coverage-gap list above).
7. **R10 vs Task 2's M4: drop the per-key rows.** Confirmed: `doctor`
   prints one line per selected source (Task 2's own gate test,
   `aae338c`), and the security review's finding 11 (M4 partial) is
   about *naming the offending key on that one line*, not about adding
   per-key rows — a narrower residual than the ruling anticipated,
   routed to 5c's security fold.

### The pass-end security review

`docs/research/2026-09-23-rung-5b-ii-security-review.md`:
scope `git diff 1d4ed29..HEAD -- '*.go' go.mod go.sum config/'` at
`08befea`. **No blocker.** Five majors (cross-process roster writes,
mailed links landing outside the owned prefix, a refused client
evicting a victim's per-address bucket, IPv6 per-client keys at /128,
no trusted-proxy setting) and fourteen minors.

**Disposition of the twenty 5b-i findings this review re-checked**
(the do-not-relitigate baseline for 5c and 5d). Each row states where
the fix landed (commit, file, symbol) and the test that proves it.

| Finding | Status | Landed (commit, file, symbol) | Proving test |
| --- | --- | --- | --- |
| B1 | Fixed | `f90c9ea`, `internal/auth/nonce.go` `Insert` | `TestRecordOutlivesItsToken` |
| M1 | Fixed | `f90c9ea`, `internal/identity/identity.go` `NormalizeAddress` | `TestNormalizeAddressIsAFixedPoint` |
| M2 | Fixed within one process | `f90c9ea`, `internal/auth/epoch.go` `advance` (called from `Bump`) | `TestBackwardBumpKeepsTheRevocation`; cross-process is finding 14 below |
| M3 | Fixed | `f90c9ea`, `internal/auth/token.go` `splitSigned` (strict decode + CR/LF refusal) | `TestVerifyLinkRefusals`, `TestLinkHasOneEncoding` |
| M4 | Partial | `772eb22`→`aae338c`, `cmd/dubplate-server/doctor.go` `rowStatuses`/`identitySourceStatus` | `TestDoctorCmd_NoRunReportsAllGreenOnLoadError`; the offending-key-naming half is unproven (finding 11, routed to 5c) |
| M5 | Fixed | `f90c9ea`, `internal/auth/token.go` `splitSigned` (bounded split) | `TestSeparatorFloodAllocatesLittle` |
| m1 | Fixed | `772eb22`, `internal/auth/session.go` `SessionCookie` | `TestSessionCookieAttributes` |
| m2 | Fixed on refresh, not at mint | `772eb22`, `internal/auth/session.go` `RefreshSession` | `TestRefreshClampsAtTheExactBoundary`; see finding 12 (no lifetime-vs-cap check at mint) |
| m3 | Fixed | `772eb22`, `internal/auth/store.go` `Open` | `TestOpen_RefusesWideLeaf` |
| m4 | Fixed | `772eb22`, `cmd/dubplate-server/auth.go` `validatePublicURL` | `TestValidateMagicLinkSource_PublicURL`; echo in message is finding 10 |
| m5 | Fixed | `772eb22`, `cmd/dubplate-server/auth.go` (duration threshold bounds) | `TestValidateMagicLinkSource_DurationBounds` |
| m6 | Fixed | `772eb22`, `internal/auth/store.go` (`Store.Write` removed) | Confirmed by grep (no `Write` method on `*Store` in `internal/auth/store.go`); `store_test.go`'s corrected comment names each store's real write path |
| m7 | Fixed | `772eb22`, `internal/auth/store.go` `recordPath` | `TestRecordPathRefusalNamesNoValue` |
| m8 | Fixed | `772eb22`, `internal/auth/nonce.go` `Sweep` | `TestSweepRetainsUnreadableRecord` |
| m9 | Fixed, with a stated residual | `772eb22`, `internal/auth/session.go` `ClearSessionCookie` | `TestClearSessionCookie`. **Residual**: clearing the cookie does not invalidate the signed session value, only expiry or an epoch bump does. **No pass owns this as future work**; it is accepted by design (a stateless session model, stated in `ClearSessionCookie`'s own doc comment) and recorded permanently as declared coverage gap 5 below, not deferred to 5c or 5d. |
| m10 | ASCII only | `f90c9ea`, `internal/identity/identity.go` `NormalizeAddress` | `TestNormalizeAddress` (control/space rows); non-ASCII control and format characters (NEL, CSI, RLO, NBSP, ZWSP) still pass (finding 8, routed to 5c) |
| m12 | Fixed | `772eb22`, `internal/auth/store.go` `RecordID` | `TestRecordID_KindSeparatesTheDigest` |
| m14 | Fixed | `772eb22`, `internal/auth/epoch.go` `Bump` (subscriber notification under recover) | `TestBumpRecoversFromPanickingSubscriber` |
| m11, m13 | Carried to 5d | not landed | As the plan states: the `ReadWritePaths` grant and the env template's nine keys |

**Disposition of this review's own nineteen new findings.** Geoff
accepted the conductor's recommendation on 2026-09-23 that majors 1-4
plus the relevant minors become 5c's security-fold Task 1 and major 5
become a 5d deploy gate (`docs/STATUS.md`, current "Next action"
item 1). **The granular per-minor split below is the conductor's own
ruling of 2026-09-23, made under that acceptance rather than a
separate decision Geoff read finding by finding.** Every finding
appears exactly once, including minor 6, which the previous draft
round omitted.

| Finding | Owner |
| --- | --- |
| Major 1 (cross-process roster writes) | 5c security-fold Task 1 |
| Major 2 (mailed links outside the owned prefix) | 5c security-fold Task 1 |
| Major 3 (refused client evicts a victim's per-address bucket) | 5c security-fold Task 1 |
| Major 4 (IPv6 per-client keys at /128) | 5c security-fold Task 1 |
| Major 5 (no trusted-proxy setting behind the tunnel) | 5d deploy gate |
| Minor 6 (kind-mode global ceiling enumeration and membership oracle) | 5c security-fold Task 1 (proposed here for the first time; grouped with the other oracle-shaped findings 10 and 12) |
| Minor 7 (roster loader accepts non-schema files) | 5c security-fold Task 1 |
| Minor 8 (non-ASCII control/format characters pass `NormalizeAddress`) | 5c security-fold Task 1 |
| Minor 9 (forwarded hop not validated as an IP) | 5d deploy gate |
| Minor 10 (refusal message echoes a public URL's userinfo password) | 5c security-fold Task 1 |
| Minor 11 (M4 partial: `doctor` names no offending key) | 5c security-fold Task 1 |
| Minor 12 (`doctor` green for sign-in-breaking configurations) | 5c security-fold Task 1 |
| Minor 13 (`AddMember` accepts an empty actor / invalid channel) | 5c security-fold Task 1 |
| Minor 14 (epoch monotonicity holds only within one process) | 5d, lands with the offboarding verb |
| Minor 15 (`streamer_username` uniqueness case-sensitive, admits LIKE metacharacters) | 5c security-fold Task 1 |
| Minor 16 (background-send load observable under the generic answer) | recorded residual, no fix required |
| Minor 17 (request-link charges no limiter for an address that fails to normalize) | 5c security-fold Task 1 |
| Minor 18 (origin check admits `http://` on the same host absent `Sec-Fetch-Site`) | joins 5c's served header set (HSTS) |
| Minor 19 (unreadable nonce record never swept) | 5c security-fold Task 1 |

M4's partial status is covered by finding 11 above; m10's ASCII-only
status is covered by finding 8 above.

### The declared coverage gaps

**Six declared**, against the running total of seventy-three after
5b-i, bringing the total to **seventy-nine**. One item the plan's own
expected list carried does not apply, corrected below.

1. **No page exercises any of this** (5c). The login page, the
   interstitial, and the add-member page are the first real callers of
   the request-link handler, the consumption handler, and the roster's
   write path. Confirmed: no `web/` tree exists yet in this pass.
2. **No test exercises a real bearer-authenticated bypassed surface**
   (Task 8), narrower than revision 1's wording: no bypass is
   registered at rung 5 (the security review's own "attacked and found
   sound" list confirms this), so the gap is the companion rule (a
   bypassed path may not read the session cookie) rather than a
   registered-but-untested bypass.
3. **The serve unit's `ReadWritePaths` grant does not exist** (security
   review m11), 5d's.
4. **The env template does not carry the nine keys the selected source
   requires** (security review m13), 5d's; the no-deploy carry stands.
5. **A stateless session survives sign-out** (security review m9), by
   design, permanently — see the disposition table's m9 row above.
6. **The mail transport is proved against a local test server only**
   (`internal/notify/notify_test.go`'s `startTestServer` binds
   `127.0.0.1:0`), never against a real provider.

**Corrected against the plan's expected list:** item 7 ("M5's
failing-first test, if the allocation measurement could not be made to
fail") does not apply. `TestSeparatorFloodAllocatesLittle`
(`internal/auth/token_test.go`) measures `AllocedBytesPerOp()` against
a ceiling derived from the flood's own length and passes without a
fallback, so no gap is recorded for it.

### The spec corrections this pass owes

- **R-5b18's two.** `docs/HISTORY.md`'s 5b-i review record and the R24
  erratum both say the 5b-i security read was not run and is owed to
  5b-ii's window as its first item. It ran on 2026-09-12 and is
  recorded at
  `docs/research/2026-09-12-rung-5b-i-security-review.md`.
- **R-5b26's.** The erratum landed at `b91bae1`
  (`docs/specs/2026-09-08-rung-5-identity-auth.md:2287-2296`)
  states that R33's redirect validation "landed in `internal/auth` at
  5b-i". No such function exists; Task 0 item 10 and this close both
  prove it by grep. The validator lands with rung 12's parameter or not
  at all.
- **The ceiling and the reservation.** 5b-ii ran at 7.5M rather than
  the 4.5M provisional 5b-i's plan carried; rung 5 stands at 24.4M
  against the live 15.75M reservation (`spec:2042-2050`, the erratum
  that moved the reservation from 14.5M to 15.75M), the 8.65M
  difference recorded against the reservation rather than taken from
  another pass.
- **`spec:1587-1593`'s threshold list against Task 3's declarations
  agrees**: ten thresholds named there (session lifetime, session
  absolute cap, link expiry, link requests/hour/address, link
  requests/hour/client, consumption attempts/hour/client, the global
  outbound ceiling, the auth request body cap, the nonce retention
  floor, and the unknown-address behavior), four landed at 5b-i (Task
  4 of that pass) and six here (Task 3). No correction needed; recorded
  as confirmed rather than corrected, since revision 2.1 asked for the
  check.
- **`spec:2332`'s 5d row** assigns "the example roster file (R37)" to
  5d, but this pass creates `config/roster.example.json` at Task 4
  (R-5b20), because Task 4's loader test is what keeps it from rotting.
  The 5d row should drop the example-file clause.
- **The launch rulings' own item (c): 5b-i `HISTORY:` pointers.** Every
  one of the following 5b-i `HISTORY:` pointers in this plan sits
  **exactly** 47 lines low at HEAD (Task 0 item 18, confirmed
  independently by this close); the "47 lines low" claim below covers
  only this set, not every `HISTORY:` pointer the plan carries:
  `:496-530` re-keys to `:543-577` (the 5b-i to 5b-ii handoff),
  `:531-544` re-keys to `:578-591` (the re-derivation basis paragraph
  that plan:465-469 cites, this close's own re-check catching the
  pointer the prior draft round omitted), `:533-535` to `:580-582` and
  `:540-543` to `:587-590` (both budgets), `:411-416` to `:458-463`
  (what this pass owes / the plan-prep tail).
  **Correction to this section's own prior claim**: the prior draft
  round said `:1692` (the 5a-flag citation the plan makes at `:488`,
  "`HISTORY:1688`") "resolves correctly at its cited line and needs no
  correction". It does not: `docs/HISTORY.md:1692` is 5a's `strip.go`
  header-list line, not a spend figure. The two 5a pointers this plan
  actually uses (`plan:487-488`) are `HISTORY:1688` and `HISTORY:1692`
  in the plan's own text; at HEAD they resolve to **`:1735`** (5a's
  Task 0 spend, 0.13M) and **`:1739`** (5a's 3.4M flag), each 47 lines
  low like every other pointer in this set, and both are in scope for
  launch ruling (c) even though the ruling's own item 18 named only the
  5b-i-labeled ones.
  **Two more pointers do not fit the uniform 47-line shift and are
  recorded separately.** `plan:347` cites `HISTORY:391-392` for R-V
  (the 30-second read timeout and the `http.MaxBytesReader`
  alternative); at HEAD that line range holds unrelated Task-4 prose,
  and R-V's own sentence resolves instead to `:568-569` in the 5b-i
  entry's own carry-forward list and to `:1605-1607` in 5a's original
  entry — neither offset is 47. `plan:1329` cites `HISTORY:526-529` "of
  the 5a entry" for the blast-radius class 5a's one planning miss
  belonged to; that range now holds 5b-i prose, and the planning-miss
  sentence itself resolves to `:1740-1743` in 5a's own entry, again not
  a 47-line shift. Both citations were already mis-keyed at plan
  authoring rather than merely drifted since; this close records the
  correct resolution rather than re-deriving a shift that does not
  apply.
- **No other Task-0 note rose to a correction.** The `Makefile:33`
  wording variant ("The web legs rejoin", plural, present tense) that
  Task 0 item 21 flagged is not a defect in this plan; it is carried
  into the 5c handoff below as the exact wording 5c's correction should
  expect.

### The 5c handoff

5c inherits rather than rebuilds: the roster's interface and its write
path, which the `member` verb and the add-member page both call; the
store's write-open constructor (`CreateRoster`), which `member add
--admin` calls to create the first roster and which refuses to create a
second, so 5c inherits a call rather than reopening the package; the
single resolution point (`normalizeMemberAddress`), which 5c's pages
must not bypass; **the admin flag's channel refusal, enforced in the
write path** (`AddMember` refuses `Admin` on every channel, a service
token included, `internal/auth/roster.go:323-331`), so the add-member
handler inherits a refusal rather than a rule to remember, and 5c's
CLI `member add --admin` brings the admin grant together with its own
caller rather than relying on any permitted channel; the verify-identity seam
and its channel, which the add-member handler reads the admin flag
through (never the cookie); the origin check, which covers 5c's
add-member POST by construction, not by anything the page does; fork
4's kind answer as **data**, with the copy 5c's to write and the admin
contact coming from the roster's admin entry; the consumption route's
`no-store` and `Referrer-Policy: no-referrer` obligations, which are
5c's served header set (joined by HSTS, security review minor 18); the
interstitial's contract, that it reads the token from
`window.location`, clears it from the address bar, and POSTs it in a
form field; the logout copy's bound, that a stateless session survives
sign-out permanently (declared gap 5, not deferred work); the
security-fold Task 1 (the conductor's ruling, under Geoff's
2026-09-23 acceptance of the majors-and-minors split, that routes the
security review's majors 1-4 and minors 6, 7, 8, 10, 11, 12, 13, 15,
17, 19 there, above); and the three rejoin-sentence corrections
(the Gate section's "The `web/` leg rejoins at rung 7" sentence, at
**`CLAUDE.md:140`** — `scripts/check.sh:5`, `Makefile:33`, the last
worded "The web legs rejoin", plural, present tense),
`scripts/deploy.sh`'s `config/contributors.yaml` delivery left alone by
R37.

**Also carried to 5c** (from STATUS, unchanged by this close): rename
`ErrRosterAdminChannelForbidden` when the CLI admin grant lands; the
last-admin lockout guard, atomic against concurrent self-demotion; the
session-refresh call site (the seam's same-origin condition plus the
clamped `RefreshSession`, since no 5b-ii handler verifies a session);
and the magic-link mail copy (a minimal placeholder today).

**Carried to 5d** (unchanged): the spec's "`doctor` exits 0" acceptance
row must run after the bootstrap/import verb; the serve unit's
`ReadWritePaths` grant; the env template's nine missing keys (the
no-deploy carry); the trusted-proxy setting and its minors 9 and 14.

### Both budgets

Tokens: about **6.4M against the 7.5M ceiling**, the 6.0M flag crossed
at the pass-end security read. Attended time: zero planning misses (no
question surfaced after launch that a planning question would have
caught) and zero execution sittings, except the conductor's own
mid-plan launch rulings (five, ruled together before Task 1 dispatched,
which the plan itself states do not count as a sitting). Gate: `bash
scripts/check.sh` measured **1m10.099s real** at this close's own
re-run (see "The gate evidence" above).

---

*End of draft. Corrections above are drafted for the conductor's fold;
none has been applied to `docs/HISTORY.md`, `docs/STATUS.md`, or the
spec.*
