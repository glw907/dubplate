# dubplate history (born musicbox; entries keep the names of their day)

Per-pass ledger, newest first.

## Rung 5, pass 5b-ii (the roster, the mail seam, the rate limiters, the origin check, the auth handlers, the security fold): closed 2026-09-23, lane 1

Executed against `docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md`
(revision 2.1, launch-ruled 2026-09-23) in the main tree, from `1d4ed29`
(5b-i's merge). Task 0 (`opus`, fresh context) ran at `f75b0d9`
(`research/handoffs/2026-09-12-5b-ii-task0-staleness.md`); the Go tree
was unchanged at `dfda0ff`, so its evidence stands unrefreshed through
the close. Its one live finding, that every 5b-i `HISTORY:` pointer in
this plan sits 47 lines low, is applied as a dated note in the plan's
header (launch ruling c) rather than by rewriting each pointer. The pass
ran in the main tree on master, so its close has no merge: the ledgers
fold in one commit (`docs: fold the 5b-ii close into the ledgers`).

Tasks 1 through 11 ran across four segments: A (Tasks 1-3, closed
`aae338c`), B (Tasks 4-6, closed `dfa7155`/`3ca0550`), C (Tasks 7-8,
closed `c09bf44`, simplifier `105ad50` + `95e7607`), and D (Tasks 9-11,
closed `c2613ba`, simplifier `08befea`), followed by the pass-end
`web-auth-security-reviewer` (`46e37d1`) and this close (Task 12). The
conductor reordered segment D2 on 2026-09-23 so the security read ran
before Task 12, letting the close record the read and any fix it forced.

### The gate evidence

`go test ./...` at the merged tip: **2m23.8s** wall (`internal/harness/conform`
at 140.9s and `internal/harness` at 11.6s dominate; every other package
under 2.2s). Package-scoped: `internal/auth` **2.1s**, `internal/identity`
**0.006s**, `internal/notify` **0.027s**, `internal/serve` **1.1s**
(all `-count=1`). `bash scripts/check.sh` ran green (exit 0) through
`cairn-run-gate` at Task 12's close. The wrapper printed no wall-clock
total at the time, and a bare "green" is the claim 5b-i's close reviewer
escalated, so the close re-ran the same gate at its own commit with
`time CAIRN_GATE_LANE=light cairn-run-gate 'bash scripts/check.sh'`:
**1m10.099s real** (`6m35.341s` user, `0m3.586s` sys) for a clean pass.

One process defect surfaced in execution. The segment B run committed
Task 5 on top of Task 4 while Task 4's review escalation was still open,
because a reviewer's `escalate` did not stop a sequential `pass-execute`
run. The conductor ruled both tasks in one fix round (`dfa7155`), and
the workstation's `pass-execute.js` gained a `stopOnEscalate` option that
ends a sequential run after any task whose final status is not accepted
(`~/.dotfiles` `ce70b5f`, 2026-09-23).
That is a fresh measurement, not a reconstruction of an earlier wall
time no artifact recorded.

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
(`internal/auth/roster.go:323-331`). That is the conductor's ruling that
the invariant lives in the write path, not in 5c's caller; 5c's CLI
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

**Task 1: 8 mutations, 8 fired.**

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

**Task 2: 3 mutations, 2 fired, 1 unfired (reason recorded).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | The fix round's own revert: drop the `Validate` rerun from `identitySourceStatus` | `cmd/dubplate-server/doctor.go` `identitySourceStatus` | yes | `TestDoctorCmd_NoRunReportsAllGreenOnLoadError`'s magic-link-row assertion fails |
| 2 (named 1) | Unwrap one level instead of recursing | `cmd/dubplate-server/doctor.go` `flattenErrors` | yes | `TestRowStatuses_FlattensNestedJoin` (`B_KEY` row differs) |
| 3 (named 2) | Drop the `Optional`-consultation loop, leaving the `Validate` rerun in place | `cmd/dubplate-server/doctor.go` `identitySourceStatus` | **no** | `TestDoctorCmd_ShortPreviousSecretMakesSourceUnsatisfied` still passes: the `Validate` rerun this dispatch added independently checks the one `Optional` key magic-link currently declares, so no test can currently distinguish the `Optional` branch from the `Validate` branch. The load-bearing property (no all-green row under a `Load` refusal) still holds through the other path. A future `Optional` key that `Validate` does not itself check would restore this mutation's bite. |

**Task 3: 2 mutations, 2 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Drop the `authRequestBodyCap` floor check | `cmd/dubplate-server/auth.go:313` `checkIntBounds` | yes | `TestValidateMagicLinkSource_RateAndCapBounds` (4 floor subtests) |
| 2 | Widen the unknown-address accepted set to also accept `"silent"` | `cmd/dubplate-server/auth.go:279` `validateMagicLinkSource` | yes | `TestValidateMagicLinkSource_UnknownAddressAnswer`/`unrecognized` |

**Task 4: 8 mutations, 8 fired.**

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

**Task 5: 7 mutations, 6 fired, 1 unfired (reason recorded).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | Write before validating | `AddMember` (dropped early returns) | yes | `TestAddMember_ByteIdentityOnFailure` (3 of 5 subtests) |
| 2 | Take the lock after the re-check | `AddMember` (moved `r.mu.Lock()`) | yes | `TestAddMember_ConcurrentStreamerUsernameCollision`, `-race -count=20` |
| 3 | Use `atomicfile.WriteFile` rather than `WriteStream` | `AddMember` | **no** | `go test -race ./internal/auth/...` stayed green; per the plan this is a review-caught defect (the durability doc comment stops matching the call), not a test-caught one, and no test was written to force it |
| 4 | Skip the re-read | `AddMember` (dropped the post-write reparse) | yes | `TestAddMember_WriteReReadAndMode`, `TestAddMember_ConcurrentDistinctMembersBothSurvive`, `TestAddMember_ConcurrentStreamerUsernameCollision` |
| 5 | Drop the channel from the audit record | `AddMember` (logger call) | yes | `TestAddMember_AuditRecord` |
| 6 | Accept the admin flag on any assertion channel | `AddMember` (guard forced `if false`) | yes | `TestAddMember_ByteIdentityOnFailure` (cookie-, header-, and service-token-channel admin rows) |
| 7 | Let the read-open tolerate an absent file | `OpenRoster` | yes | `TestAddMember_WriteReReadAndMode` |

**Task 6: 4 mutations, 4 fired.**

| # | Mutation | Site | Fired | Proving test |
| --- | --- | --- | --- | --- |
| 1 | Guard the recipient and not the subject | `internal/notify/notify.go` `Send` | yes | `TestSend_HeaderInjectionGuard`/subject CR and LF |
| 2 | Guard only LF, not CR | `internal/notify/notify.go` `guardHeaderValue` | yes | `TestSend_HeaderInjectionGuard`/recipient CR, subject CR |
| 3 | Guard after composing the message | `internal/notify/notify.go` `Send` (guard-vs-compose order) | yes | `TestSend_HeaderInjectionGuard`/recipient CR and LF |
| 4 | Route port 465 through STARTTLS | `internal/notify/notify.go` `Send` (TLS branch selection) | yes | `TestSend_ImplicitTLS465` |

**Task 7: 6 mutations, 5 fired, 1 unfired (deferred to Task 10's own ledger).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| M1 | Remove the map's bound | `internal/auth/limit.go` `keyedLimiter.allow` | yes | `TestKeyedLimiterBoundHoldsUnderFlood`, `TestKeyedLimiterEvictsLeastRecentlyUsed`, `TestKeyedLimiterConcurrent` |
| M2 | Share one `keyedLimiter` across all three limiters | `internal/auth/limit.go` `NewLimiter` | yes | `TestKeySpacesDoNotShareBuckets` (both subtests); 3 more test functions, 5 subtests total |
| M3 | Evict most-recently-used instead of least | `internal/auth/limit.go` `keyedLimiter.allow` | yes | `TestKeyedLimiterEvictsLeastRecentlyUsed`, `TestKeyedLimiterConcurrent` |
| M4 | Key the per-client limiter on the wrong source | call site in Task 10 | **no (here)** | No test in this package can reach it: `Limiter` takes the client address as a string and derives none, so this is a call-site defect. The plan routes this mutation to Task 10, where it is proved (see Task 10 row 12). |
| M5a | Only the address key gates | `internal/auth/limit.go` `AllowLinkRequest` | yes | `TestLinkRequestKeysTripIndependently`/`many_addresses_from_one_client`, `TestKeySpacesDoNotShareBuckets`/`link requests per client` |
| M5b | Only the client key gates | `internal/auth/limit.go` `AllowLinkRequest` | yes | `TestLinkRequestKeysTripIndependently`/`one_address_from_many_clients`, `TestLinkRequestNormalizesAddress` |

**Task 8: 7 mutations, 6 fired, 1 unfired (reviewer-graded, no test can reach).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | Drop the tightening; use the stdlib default | `internal/serve/origin.go` `allowed()` | yes | `TestOriginCheck`, `TestOriginCheckBearerBypass`, `TestOriginCheckCoversSubtreeAndFallthrough`, `TestOriginCheckRunsAfterNormalize`, `TestOriginRefusalEmitsTheRefusalRecord` (5 top-level, 6 subtests) |
| 2 | Move the check inside the prefix's routed handler | `internal/serve/serve.go` `Handler()` | yes | `TestOriginCheckCoversSubtreeAndFallthrough`, `TestOriginRefusalEmitsTheRefusalRecord` |
| 3 | Move the check outside `Normalize` | `internal/serve/serve.go` `Handler()` | yes | `TestOriginCheckRunsAfterNormalize`, `TestEchoMethodsRefusedBeforeRouting` (4 rows) |
| 4 | Let a bypassed path read the session cookie | contract on `originCheck`'s doc comment | **no** | No test can reach it: no bearer-authenticated path exists yet at this rung. Graded by a reviewer against the doc comment, per the plan's own instruction. |
| 5 | Consult the bypass set after the neither-header denial | `internal/serve/origin.go` `allowed()` | yes | `TestOriginCheckBearerBypass`/`bypassed path, POST carrying neither header, is allowed` |
| 6 | Register through `AddInsecureBypassPattern` alone | `internal/serve/origin.go` `addBearerBypass()`/`allowed()` | yes | `TestOriginCheckBearerBypass`/same subtest |
| 7 (extra, not named in the plan) | Drop the `AddInsecureBypassPattern` call | `internal/serve/origin.go` `addBearerBypass()` | yes | `TestOriginCheckBearerBypass`/`bypassed path, POST carrying a cross-origin Origin, is allowed` |

**Task 9: mutation count and outcomes not recoverable from the durable record.**

The plan names 4 mutations for Task 9 (`plan:2659-2663`): reading the
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

**Task 10: 12 mutations, 12 fired.**

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

**Task 11: 4 mutations, 3 fired, 1 structurally unreachable (reason recorded).**

| # | Mutation | Site | Fired | Proving test / reason |
| --- | --- | --- | --- | --- |
| 1 | Mount the handlers outside the owned prefix | `cmd/dubplate-server/serve.go` `newServeShell` | yes | `TestNewServeShell_MountsTheThreeAuthRoutes`/`consume-link` |
| 2 | Start the sweep driver without the verb's context | `cmd/dubplate-server/serve.go` `newServeShell` | yes | `TestNewServeShell_SweepDriverStopsWithTheGivenContext` (added specifically because the pre-existing `TestStartNonceSweep_RunsAndStopsWithContext` calls `startNonceSweep` directly and did not catch this on the first pass) |
| 3 | Open the stores after listening | `cmd/dubplate-server/serve.go` `newServeShell` (`auth.OpenRoster` err check) | yes | `TestNewServeShell_MissingRosterRefusesBeforeConstruction` |
| 4 | Register a bypass pattern for the fallthrough | none in this task's Files | **no** | Structurally unreachable: `internal/serve.Shell` exposes no exported method to register a bearer bypass (`originCheck.addBearerBypass` is unexported on an unexported type), and this task's Files (`cmd/dubplate-server/serve.go`, `internal/auth/nonce.go`) hold no code path that could register one; `internal/serve/origin.go` is outside this task's Files. `TestNewServeShell_StateChangingFallthroughRequestIsDenied` proves the absence behaviorally, but the mutation cannot be introduced from this task's own surface. |

### The twelve resolutions (R-5b17 through R-5b28)

**The plan's grep-recoverable-marker mechanism did not hold for this
pass.** The plan builds this table from `grep -rn 'ruled, conductor
2026-09-12, R-5b'` (`plan:701-703`, `:3009-3013`). That grep, and the
looser `grep -rn 'R-5b1[7-9]\|R-5b2[0-8]'`, together find markers for **3 of
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
   (see the corrected coverage-gap list below).
7. **R10 vs Task 2's M4: drop the per-key rows.** Confirmed: `doctor`
   prints one line per selected source (Task 2's own gate test,
   `aae338c`), and the security review's finding 11 (M4 partial) is
   about *naming the offending key on that one line*, not about adding
   per-key rows. That is a narrower residual than the ruling anticipated,
   routed to 5c's security fold.

### The pass-end security review

`docs/superpowers/research/2026-09-23-rung-5b-ii-security-review.md`:
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
become a 5d deploy gate (`docs/STATUS.md` "Next action"
item 1, at `2a416b5`). **The granular per-minor split below is the conductor's own
ruling of 2026-09-23, made under that acceptance rather than a
separate decision Geoff read finding by finding.** Every finding
appears exactly once.

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
   design, permanently; see the disposition table's m9 row above.
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

The close applied the first, second, third, and fifth items below as
dated errata (2026-09-23, 5b-ii close) under the rung 5 spec's "Errata
the close applies", with a fifth erratum re-keying that section's
`CLAUDE.md:118` pointer to `CLAUDE.md:140`. The pointer re-keys are one
dated note in the plan's header.

- **R-5b18's two.** `docs/HISTORY.md`'s 5b-i record (the "security
  cycle did not run this window" bullet in the 5b-i entry) and the R24
  erratum both say the 5b-i security read was not run and is owed to
  5b-ii's window as its first item. It ran on 2026-09-12 and is
  recorded at
  `docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`.
  The 5b-i entry stays as written; this entry is its correction.
- **R-5b26's.** The erratum landed at `b91bae1`
  (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md:2287-2296`)
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
- **`spec:1587-1593`'s threshold list agrees with Task 3's
  declarations**: ten thresholds named there (session lifetime, session
  absolute cap, link expiry, link requests/hour/address, link
  requests/hour/client, consumption attempts/hour/client, the global
  outbound ceiling, the auth request body cap, the nonce retention
  floor, and the unknown-address behavior), four landed at 5b-i (Task
  4 of that pass) and six here (Task 3). No correction needed; recorded
  as confirmed, since revision 2.1 asked for the check.
- **The pass-split table's 5d row** assigns "the example roster file
  (R37)" to 5d, but this pass creates `config/roster.example.json` at
  Task 4 (R-5b20), because Task 4's loader test is what keeps it from
  rotting. The 5d row drops the example-file clause.
- **The launch rulings' own item (c): the plan's `HISTORY:` pointers.**
  At `2a416b5`, before this entry landed, every 5b-i `HISTORY:` pointer
  in the plan sat exactly 47 lines low (Task 0 item 18, confirmed
  independently by this close): `:496-530` resolved to `:543-577` (the
  5b-i to 5b-ii handoff), `:531-544` to `:578-591` (5b-i's "Both
  budgets", which `plan:480-484` cites), `:533-535` to
  `:580-582` and `:540-543` to `:587-590` (the two spans quoted from
  it). The two 5a pointers shift the same way: `plan:503`'s
  `HISTORY:1688` resolved to `:1735` (5a's Task 0 spend, 0.13M) and
  `plan:512`'s `HISTORY:1692` to `:1739` (5a's 3.4M flag). Three
  pointers do not fit the uniform shift. `plan:978`'s `:411-416`,
  the "not run this window" record, resolved to the two bullets at
  `:456-462`. `plan:362` cites `HISTORY:391-392` for R-V (the
  30-second read timeout and the `http.MaxBytesReader` alternative); at
  `2a416b5` R-V's sentence sat at `:568-569` in the 5b-i entry's
  carry-forward list and at `:1605-1607` in 5a's original entry.
  `plan:1344` cites `HISTORY:526-529` "of the 5a entry" for the
  blast-radius class of 5a's one planning miss; that sentence sat at
  `:1740-1743`. Those last two were mis-keyed at plan authoring rather
  than drifted since. The plan's header note gives every one of these
  re-keyed against the file with this entry in place.
- **No other Task 0 note rose to a correction.** The `Makefile:33`
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
17, 19 there, above); the three rejoin-sentence corrections, at
**`CLAUDE.md:140`** (the Gate section's "The `web/` leg rejoins at
rung 7"), `scripts/check.sh:5`, and `Makefile:33` (worded "The web legs
rejoin", plural, present tense); and `scripts/deploy.sh`'s
`config/contributors.yaml` delivery, left alone by R37.

**Also carried to 5c** (from STATUS, unchanged by this close): rename
`ErrRosterAdminChannelForbidden` when the CLI admin grant lands; the
last-admin lockout guard, atomic against concurrent self-demotion; the
session-refresh call site (the seam's same-origin condition plus the
clamped `RefreshSession`, since no 5b-ii handler verifies a session);
and the magic-link mail copy (a minimal placeholder today).

**Carried to 5d** (unchanged): the spec's "`doctor` exits 0" acceptance
row (the pass-split table's 5d exit gate) must run after the
bootstrap/import verb; the serve unit's
`ReadWritePaths` grant; the env template's nine missing keys (the
no-deploy carry); the trusted-proxy setting and its minors 9 and 14.

### Both budgets

Tokens: about **6.4M against the 7.5M ceiling**, the 6.0M flag crossed
at the pass-end security read. Attended time: zero planning misses (no
question surfaced after launch that a planning question would have
caught) and **one execution sitting**: the combined question at the
6.0M flag (fold the security majors here or as 5c's Task 1), which Geoff
answered with the conductor's recommendation. The conductor's launch
rulings before Task 1 do not count, as the plan states. Gate: `bash
scripts/check.sh` measured **1m10.099s real** at this close's own
re-run (see "The gate evidence" above).

## Rung 5, pass 5b-i (the auth core: config contract, token, session, state records): closed 2026-09-12, lane 2

### Process changes applied 2026-09-12

The process improvement plan (`research/2026-09-12-process-improvement-plan.md`,
revision 2.1) closed with six owner-owed decisions, all ruled by Geoff on
2026-09-12 as applied: the merge ritual's cadence stays boundary-only rather
than a full-gate CI job on every push; the in-tree merge gate gains a
branch-history `gitleaks` leg; `pass-execute.js` and `diff-reviewer.md` stay
shared workstation assets, gaining optional fields and an opt-in argument
rather than a dubplate fork; the close's authorship moves to one fold agent
with an independent `diff-reviewer` read over its diff; the plan-prep
fan-out takes the contract-and-criteria, mechanics-and-feasibility, and
domain-risk three-lens default; and the merge ritual's cross-lane review
fires on a disjunction adding the shared-package case to the existing
non-fast-forward case.

The fresh-context Go idiom and architecture read
(`research/2026-09-12-go-idiom-architecture-read.md`, `979db72`) graded the
greenfield tree "sound with nits": correct, well tested, and mechanically
clean, architecturally sound wherever a real consumer exists, with three
recurring defects (a large speculative exported surface, doc comments citing
the plan process rather than the code, and near-duplicate bodies written by
different tasks) all traced to one cause, that the unit of review is a
task's diff rather than a whole package. It argued for three process
changes, all ruled and landed: a per-rung whole-package architecture read
run with no plan in context (the `go-architecture-reader` agent, `1c5c76d`);
a "no consumer, no export" gate leg (the exported-unused check below); and a
gate leg against process citations in shipped comments paired with a
mechanical simplifier brief (the citation check below and
`.claude/instructions/simplifier-brief.md`, `1c5c76d`).

The gate forms landed at `2a81a0c`: `--scope-changed`, `--reduced`, and
`--merge`, with both staged legs scoped by a git-derived change set rather
than an argument. The architecture reader, the simplifier brief, the
`close-evidence` tool, and three ruled `CLAUDE.md` sentences landed at
`1c5c76d`. The two staged legs (313 process-citation lines, mostly
`ruled, conductor` and `Resolution R` trailers; 134 exported-unused symbols,
mostly packages whose consuming rung is later, so the rule exempts
`internal/harness/**` and exempts any other package until its consumer
lands) warn today; the Go read's ten findings and the `make close-evidence`
target are owed to the comment-cleanup task that flips them to fail.
`~/.dotfiles`'s `pass-execute.js` gained the same optional fields and
opt-in argument at `d2c4831`, the shared-asset half of owner-owed decision 3.

Recorded reading: the twelve conductor-owed folds were ruled by
recommendation on "then start" (2026-09-12 00:44); the ratify-or-reverse
table is in `plans/2026-09-12-overnight-window.md`.

Executed against
`docs/superpowers/plans/2026-09-08-rung-5b-auth-core.md` (revision 2.1)
in the worktree at `.claude/worktrees/lane-2-5b`, branch `lane-2-5b`
from `7ba3cc1`, merged to master at `1d4ed29`. Task 0 (`opus`, fresh
context) found no blockers beyond the pointer drift the dispatch notes
had already corrected. Tasks 1 through 9 ran across three segments:
2A (`wf_61f3cf53-665`, Tasks 1 through 3, closed `14fad2c`), 2B
(`wf_769302f4-cef`, Tasks 4 through 7, closed `8888cfe`), and 2C
(`wf_be7f90c5-281`, Tasks 8 through 10, closed `b832f8a`). This entry
is Task 10's close, folded by the conductor.

### The gate evidence

The worktree's full `bash scripts/check.sh` ran green in 3m20s before
the merge; master's ran green in 3m44s after it. `go test ./...`
measured 2m20s at Task 0's baseline. `go test ./internal/auth/...
./internal/identity/... -count=1` at the merged tip runs green in 0.7s
wall (`internal/auth` 0.028s, `internal/identity` 0.005s), which is the
whole of this pass's own test surface: the module's minutes are the
corpus and shadow suites, not the auth core.

Two structural confirmations the close verified rather than asserted.
`git diff 6c745c5..1d4ed29 -- internal/serve/` is empty: 5b-i wrote no
`internal/serve` file, and `cmd/dubplate-server/serve.go` is the one
wiring change (Task 6), which is R-5b15 discharged by measurement.
`go.mod` and `go.sum` are unchanged across the same range: the two
module dependencies R26 budgets are still 5b-ii's to take.

Seven test files that existed before this pass changed, each for a
stated reason, and no existing assertion was weakened:

| Existing test file | Why it changed |
| --- | --- |
| `internal/config/config_test.go` | `Load` became a generic requirement over every owned-now registered row (A1), so the load tests moved onto registries rather than four hand-written names, and the identity-source resolution rows are new |
| `internal/config/registry_test.go` | the `IdentitySource` declaration family joined the registry (A2), and the collision gate gained its six identity-source rows |
| `internal/config/keys_test.go` | `keys.go`'s pending "identity source" row and its "auth and identity" reserved class were deleted in Task 3, so the tests asserting them went with them |
| `cmd/dubplate-server/templatekeys_test.go` | `DUBPLATE_IDENTITY_SOURCE` became an owned-now row with a literal template line (R-5b7), which both template tests now cover against the production registry |
| `cmd/dubplate-server/doctor_test.go` | the doctor-exit assertion moved off scratch registries onto `productionRegistry()` (Task 3's fix round), where a wrongly declared real name can fail it |
| `cmd/dubplate-server/check_test.go` | the check verb's config load now resolves an identity source, so its fixture environment carries the source's required names |
| `cmd/dubplate-server/serve_test.go` | the serve verb builds its shell through `newServeShell` and passes the real cookie name to `NewStrip` (Task 6), retiring 5a's declared coverage gap 4 |

### What landed

Task 1 (`13e543f`, `d090eed`) added `internal/identity`: a three-value
`Channel`, an `Identity` keeping `Principal` distinct from `Address`
(2.0 hook 5's first half), and `NormalizeAddress`, the one function
every keying, logging, and rate-limit lookup on a login address must
call.

Task 2 (`6cb2656`, `opus`) changed `internal/config/config.go` and
`registry.go`: `Load` now requires every owned-now wall row the
registry carries rather than four names written by hand, closing risk
11 for every future rung by construction (A1). The identity-source
family is the wall's sixth classification: a rung declares a source by
its selector key, required keys, optional keys, and a validation hook
receiving the source's resolved values plus the `os.Environ()`
snapshot (A2, review B2, conductor-owed as applied), which is what
lets `DUBPLATE_NONCE_RETENTION` and `DUBPLATE_MAGIC_LINK_EXPIRY`
validate their own relation without `internal/config` learning a rung
5 name.

Task 3 (`ee79b5b`, fix `14fad2c`) added `cmd/dubplate-server/auth.go`:
`registerAuthConfig` declares `DUBPLATE_IDENTITY_SOURCE` as an
owned-now row and the `magic-link` source's nine required and two
optional names, retiring `internal/config/keys.go`'s pending "identity
source" row and the whole "auth and identity" reserved class in the
same commit (A3), the removal half of the rung 1b registry contract's
second exercise. Three supersession messages were rewritten:
`DUBPLATE_CONTRIBUTORS_FILE` now names `DUBPLATE_ROSTER_FILE`, and the
two Access rows describe the forward-auth seam generically. `main.go`
gained `newProductionRegistry`, the binary's one registry constructor.
The task's own handoff (`research/handoffs/2026-09-12-rung-5b-task-3.md`,
committed at `bf9fbc4`) records its mutation results in full and is the
source for this entry's Task 3 row below.

Task 4 (`5f4d315`) added the four thresholds `DUBPLATE_MAGIC_LINK_EXPIRY`
(`15m`), `DUBPLATE_SESSION_LIFETIME` (`2160h`),
`DUBPLATE_SESSION_ABSOLUTE_CAP` (`4320h`), and
`DUBPLATE_NONCE_RETENTION` (`24h`), and `validateMagicLinkSource`: the
32-byte secret floor (R-5b8), the public URL's three conditions with a
whole-loopback-set carve-out (fold 8, ruled 2026-09-12) rather than the
literal name `localhost`, and the retention-vs-expiry relation (R31).

Task 5 (`c230bf9`, `opus`) added `internal/auth/token.go`: the
magic-link token, carrying the normalized address, expiry, mint time,
and a 128-bit `crypto/rand` nonce. Every field is base64url-encoded
before the length-prefixed MAC input's "." join (A4 fold 3, ruled
2026-09-12 as a correctness fix: raw joining breaks any address whose
local part carries a dot). Verification order is fixed: MAC through
`crypto/subtle.ConstantTimeCompare`, then expiry, then the nonce, so an
unauthenticated caller cannot fill the nonce store. The previous secret
verifies and never mints. The link is built from the configured public
origin, never a request value, with the token in the query string
(R-5b13).

Task 6 (`06505a3`, `opus`) added `internal/auth/session.go`: the
`__Host-dubplate_session` cookie (R-5b9), same wire shape as the token
under its own domain-separation label. Rolling refresh moves the issue
time and the expiry, never the anchor, so the absolute cap survives any
number of refreshes (A5). The serve verb now builds its shell through
`newServeShell`, passing the real cookie name to `NewStrip`, retiring
5a's declared coverage gap 4.

Task 7 (`8888cfe`) added `internal/auth/store.go`: the shared
substrate both state records sit on, one leaf (`auth`) with one
subdirectory per record kind, an opener creating both at 0700,
`RecordID` (full SHA-256 hex, untruncated, R-5b12), an atomic write at
0600, and `validID` refusing empty, `.`, `..`, and separator-bearing
ids. A confinement test writes through every entry point, including for
an address whose local part carries `..`, and walks the filesystem to
prove nothing escapes root and no raw address appears in a filename
(A6).

Task 8 (`41e07d2`, `opus`) added `internal/auth/nonce.go`: the
create-exclusive consumed-nonce store. The insert opens each record
with `O_CREATE|O_EXCL`, syncs the record and its directory, and reports
`ErrNonceReplayed` on collision (R11); `atomicfile`'s rename-over shape
cannot carry this write because it would silently replace the proof of
consumption. Each record carries consumed-at plus the resolved
`DUBPLATE_NONCE_RETENTION`, never the token's own expiry (conductor-owed
as applied), and a record the sweep cannot parse is retained and
counted until its mtime passes retention rather than deleted on sight
(conductor-owed as applied), because deleting a crash-truncated record
would make its token consumable a second time. The sweep runs at open
and is offered as a method; the interval driver is 5b-ii's (R-5b14).

Task 9 (`8a9ee33`, then `46ce4e7`) added `internal/auth/epoch.go`: one
record per normalized address holding the epoch a bump last set,
written through `atomicfile.WriteStream` so the overwrite fsyncs the
containing directory (R11). `CheckSession` refuses a session whose
**anchor** predates its address's epoch, and `CheckLink` refuses a link
whose mint time predates it (conductor-owed as applied: the anchor
rather than the issue time, because refresh moves the issue time and a
session refreshed once after a bump would otherwise be permitted
forever, R30's own failure mode reproduced on every request). `Read` is
the authority a subscriber that missed a notification can still
consult (A8). The `46ce4e7` follow-up fixed a real gap the task's own
review caught: both checks keyed the epoch lookup on the credential's
raw address rather than the normalized one, so a session or link whose
address differed only in case silently escaped revocation; both now run
through `identity.NormalizeAddress` before keying, matching `Bump` and
`Read`, with `TestCheckSessionNormalizesAddress` made table-driven over
both bump/check case-mismatch directions.

### The three simplifier rounds

Exactly three `refactor(...)` commits sit on this pass's range, one per
segment rather than one per task: `8ae24cd` (segment 2A, tidying rung
5b's identity surface in `internal/config` for stdlib idiom),
`7d1f137` (segment 2B, folding the token's and session's
near-identical wire-shape helpers into three shared functions in
`internal/auth/token.go`, `encodeSigned`, `splitSigned`, and
`macMatches`, so the encoder, the decoder, and the constant-time
comparison cannot drift apart between the two credentials), and
`1d4ed29` (segment 2C, giving the subdirectory join one home and naming
two silences). Those are the rebased hashes on master; in the lane
branch's own history they were `52b6a20`, `a97245e`, and `ca77dcd`. Any
later reading that expects one simplifier commit per task will not find
one, and that is the deviation recorded below rather than a missing
round.

### The exported surface

`Channel`, `Identity`, `NormalizeAddress` (`internal/identity`);
`Load`'s generic owned-row requirement, `IdentitySource`,
`Config`'s active-source accessor (`internal/config/config.go`,
`registry.go`); `registerAuthConfig`, `magicLinkSource`,
`validateMagicLinkSource`, the four thresholds
(`cmd/dubplate-server/auth.go`); `LinkToken`, `MintLink`, `VerifyLink`,
`LinkURL`, `encodeSigned`/`splitSigned`/`macMatches`
(`internal/auth/token.go`); `SessionCookieName`, `MintSession`,
`VerifySession`, `RefreshSession` (`internal/auth/session.go`);
`StoreDir`, `Open`, `RecordID`, `validID`, atomic write
(`internal/auth/store.go`); `NonceStore`, `OpenNonceStore`, `Insert`,
`Sweep`, `NonceSweep` (`internal/auth/nonce.go`); `EpochStore`,
`OpenEpochStore`, `Bump`, `Read`, `Subscribe`, `CheckSession`,
`CheckLink` (`internal/auth/epoch.go`).

### What the gate caught

Task 3 ran a fix round (`14fad2c`) closing a real gap its own review
found: the doctor-exit and selector-name assertions had run only
against scratch registries, so neither could fail if the real
`DUBPLATE_IDENTITY_SOURCE` name were declared wrongly; both moved onto
`productionRegistry()`. Task 9's `46ce4e7` follow-up closed the
case-normalization gap described above, found by the task's own review
rather than by a later pass. No task ran red and uncommitted this pass.

### The auth core's contract coverage, A1 through A8

| Contract | Go site | Proved by | Reach |
| --- | --- | --- | --- |
| A1 a registered owned row is a required row | `Load` (`internal/config/config.go`) | `TestLoad_RequiresEveryRegisteredOwnedNowRow` | Full |
| A2 one identity source selected by configuration, correlated set required by mechanism | `Load`'s `resolveIdentitySource` (`internal/config/config.go`), `IdentitySource` (`registry.go`) | `TestLoad_IdentitySource`, `TestCollisionGate_IdentitySource*` (six rows) | Full |
| A3 rung 5's names declared in rung 5's own file, both retirements in one commit | `registerAuthConfig` (`cmd/dubplate-server/auth.go`) | `TestTemplateKeysCoverEveryOwnedNowRow`, `TestTemplateKeysClassifiedAgainstProductionRegistry` | Full |
| A4 the magic-link token cannot be forged, shifted, replayed, or outlived | `MintLink`, `VerifyLink` (`internal/auth/token.go`) | `TestLinkGoldenVectors`, `TestMACLengthPrefixesResistFieldShifting`, `TestVerifyLinkRefusals`, `TestVerifyLinkAcceptsPreviousSecret` | Full except the constant-time comparison's timing property (declared gap 1) |
| A5 a session cannot outlive its absolute cap, rolling refresh cannot extend it | `MintSession`, `VerifySession`, `RefreshSession` (`internal/auth/session.go`) | `TestRefreshHoldsTheAnchor`, `TestRefreshedSessionRefusedAtTheAbsoluteCap`, `TestSessionAndLinkCannotCross` | Full |
| A6 both state records named by one derivation, confined to one root | `RecordID`, `Open`, `validID` (`internal/auth/store.go`) | `TestRecordID_Vector`, `TestValidID`, `TestConfinement` | Full |
| A7 a consumed nonce cannot be consumed twice, across a restart or any accepted configuration | `Insert`, `Sweep` (`internal/auth/nonce.go`) | `TestInsertRefusesReplay`, `TestInsertRefusesReplayAcrossRestart`, `TestConcurrentInsertAdmitsOne`, `TestSweepPrunesOnCarriedExpiry`, `TestSweepRetainsMalformedRecord`, `TestSweepRetainsCrashTruncatedRecord` | Full except the directory-fsync's crash-consistency property (declared gap 2) |
| A8 the revocation seam exposes a read as well as a notification | `Bump`, `Read`, `Subscribe`, `CheckSession`, `CheckLink` (`internal/auth/epoch.go`) | `TestBumpReadDurable`, `TestCheckSession`, `TestCheckSessionRefreshedAfterBumpStillRefused`, `TestCheckLink`, `TestCheckSessionNormalizesAddress`, `TestReadAuthoritativeWithoutSubscriber` | Full |

### The resolutions, as ruled

Sixteen resolutions, R-5b1 through R-5b16, ruled by the conductor at
dispatch, 2026-09-08, stand as applied except where a later fold
corrected the reading (noted per row). Six carry a code site the marker
convention stamps; the other ten are process or plan decisions with no
single line to point at, and this table says so rather than inventing
one.

| Resolution | Ruling | Landed at |
| --- | --- | --- |
| R-5b1 | The pass splits; this plan is 5b-i | Plan only, no code site |
| R-5b2 | Auth keys declared in `cmd/dubplate-server/auth.go`, not `internal/auth` or `internal/config` | `cmd/dubplate-server/auth.go`, `registerAuthConfig` |
| R-5b3 | All twelve names declared at 5b-i though six consumers arrive at 5b-ii | `cmd/dubplate-server/auth.go`, the source's required and optional lists |
| R-5b4 | `Load` requires every owned-now wall row generically, not by naming rung 5's selector | `internal/config/config.go`, `Load` |
| R-5b5 | The per-source correlated set is a registry declaration family, not a widened integration class or a hard-coded switch | `internal/config/registry.go`, `IdentitySource` |
| R-5b6 | The two rotation-overlap names are the source's optional keys, never thresholds | `cmd/dubplate-server/auth.go`, `magicLinkSource.Optional` |
| R-5b7 | `DUBPLATE_IDENTITY_SOURCE` lands as a literal row in the env template at the task that makes it owned-now | `env/dubplate-server.env.template` |
| R-5b8 | The secret length floor is 32 bytes | `cmd/dubplate-server/auth.go`, `secretLengthFloor` |
| R-5b9 | The session cookie is named `__Host-dubplate_session` | `internal/auth/session.go`, `SessionCookieName` |
| R-5b10 | Four thresholds and their defaults (`15m`, `2160h`, `4320h`, `24h`) | `cmd/dubplate-server/auth.go`, the threshold registrations |
| R-5b11 | The state leaf is `auth`, one subdirectory per record | `internal/auth/store.go`, `leafDirName` |
| R-5b12 | The record digest is full SHA-256 hex, untruncated | `internal/auth/store.go`, `RecordID` |
| R-5b13 | The token rides in the link's query string, never a path segment | `internal/auth/token.go`, `LinkURL` |
| R-5b14 | The nonce sweep's interval is a package constant; the driver arrives at 5b-ii | **As landed, partial**: the sweep runs at open and as a callable method (`internal/auth/nonce.go`); no interval constant exists, because an unused constant fails lint and no driver calls it on a schedule at 5b-i. The plan's own text ("one hour, a package constant with its reason stated") describes 5b-ii's shape prematurely; the constant lands with 5b-ii's driver. R-5b14's operative clause ("the driver arrives at 5b-ii") is what landed |
| R-5b15 | 5b-i writes no `internal/serve` file except the binary's construction call | Confirmed by measurement: `git diff 6c745c5..1d4ed29 -- internal/serve/` is empty; `cmd/dubplate-server/serve.go` is the one wiring change (Task 6) |
| R-5b16 | A fresh-context `web-auth-security-reviewer` runs at 5b-i's own pass end, and again at 5b-ii's | **Not run this window**, for budget; owed to the next attended session as its first item (see "What this pass owes") |

### The mutation table

Named mutations as the plan predicted them per task. Task 3's row is
quoted from that task's own persisted handoff
(`research/handoffs/2026-09-12-rung-5b-task-3.md`), the only task this
pass that wrote one. The other eight tasks' per-task pass-execute
transcripts (`wf_61f3cf53-665`, `wf_769302f4-cef`, `wf_be7f90c5-281`)
are session-bound and were not persisted as repo artifacts, so this
close compiles their rows by cross-referencing the plan's named
mutations against the landed tests and doc comments rather than by
re-running each one; every named test cited below exists at the stated
site and its assertion covers the mutation's stated failure mode.

**Task 1** (5 named): folding the principal into the address field,
adding a fourth channel value, admitting the zero value, lowercasing
without parsing, and comparing the cap against rune count instead of
byte length each map to a distinct row in `identity_test.go`
(`TestIdentity_PrincipalDistinctFromAddress`,
`TestChannel_StringAndValid`, `TestIdentity_Valid`,
`TestNormalizeAddress`'s unparseable-address and multi-byte-boundary
rows).

**Task 2** (7 named): dropping the registered-row requirement, requiring
a `Pending` row, requiring the selector with no source declared,
skipping the validation hook, dropping the environment-snapshot
parameter, misclassifying a source name as unknown, and printing a
source's resolved value in `doctor` each map to
`TestLoad_RequiresEveryRegisteredOwnedNowRow`,
`TestCollisionGate_IdentitySourceCarveOut`,
`TestCollisionGate_IdentitySource`, `TestLoad_IdentitySource`'s hook
row, and `internal/config`'s doctor-output tests. Mutation 5 (dropping
the `os.Environ()` parameter) is a compile-time failure by
construction: `validateMagicLinkSource`'s signature requires it.

**Task 3** (7 named, from the task's own handoff): declaring the twelve
names without deleting the reserved class fires a `Registry.claim`
panic at registration rather than a test failure. Deleting the reserved
class without declaring the source fails
`TestTemplateKeysClassifiedAgainstProductionRegistry`, naming the five
SMTP names unknown. Clearing `Pending` instead of deleting the row
fires louder than predicted: `AddWallRow`'s own switch panics
immediately, because the row's name was never `DUBPLATE_`-prefixed, so
the plan's "Load-time demand for `identity source`" wording is
corrected by this close. Registering the selector while skipping the
template row fails `TestTemplateKeysCoverEveryOwnedNowRow`, naming
`DUBPLATE_IDENTITY_SOURCE` (R-5b7's whole argument, fired). Rewriting
`DUBPLATE_CONTRIBUTORS_FILE`'s message by deletion fails the frozen
template's own classification. Two mutations are **confirmed unfired**:
`productionRegistry` restating the register calls instead of calling
the shared constructor (a maintenance defect, not behavioral, proven by
`go test ./cmd/dubplate-server/...` staying green), and
`TestCollisionGate_IdentitySourceCarveOut` reading the production table
instead of a fixture row (passes vacuously with no pending row left to
exercise, which this task's own fixture-row rewrite closed).

**Task 4** (7 named): comparing the secret's rune count instead of byte
length, moving the floor's exclusivity, admitting plain HTTP on any
host, comparing `url.Host` instead of `url.Hostname()`, dropping the
non-empty-host condition, reversing the retention/expiry inequality,
and including the offending value in an error each map to a row in
`cmd/dubplate-server`'s auth-config validation tests (the multi-byte
secret row, the at-the-floor row, the plain-HTTP row, the
`http://localhost:8080` port row, the `example.test` row, the
equal-and-below retention rows, and the no-secret-in-errors row).

**Task 5** (7 named, `opus`): concatenating fields without length
prefixes, joining without per-field encoding, dropping the
domain-separation label, checking expiry before the MAC, returning the
nonce on failed verification, and building the link from a request host
each map to `TestMACLengthPrefixesResistFieldShifting`,
`TestLinkGoldenVectors`'s dotted-address row, `TestVerifyLinkRefusals`'s
cross-purpose and both-broken rows, and `TestLinkURL`. Comparing MACs
with `==` or `bytes.Equal` is **reported unfired by design**: no
gate-resident test can prove a timing property, and `macMatches`'s own
doc comment in `internal/auth/token.go` carries the reason
`crypto/subtle.ConstantTimeCompare` is the call (declared gap 1).

**Task 6** (6 named, `opus`): moving the anchor on refresh, dropping
`Secure` or setting `Domain`, reusing the link's domain-separation
label, checking the cap against issue time instead of anchor, moving
the anchor alongside the issue time, and leaving the serve verb passing
an empty cookie name each map to `TestRefreshHoldsTheAnchor`,
`TestSessionCookieAttributes`, `TestSessionAndLinkCannotCross`,
`TestRefreshedSessionRefusedAtTheAbsoluteCap`, and `serve_test.go`'s
cookie-name construction and fallthrough-removal rows.

**Task 7** (5 named): naming a record from its value instead of its
digest, truncating the digest, using rung 4's 0755/0644 permissions,
admitting a separator-bearing id, and joining a path in a caller
instead of through `StoreDir` each map to `TestConfinement`,
`TestRecordID_Vector`, `TestOpen_CreatesDirectoriesWithPermissions`, and
`TestValidID`. The caller-joins-its-own-path mutation is **reported
unfired by design**: nothing fails today because no second caller
exists yet to diverge. `StoreDir`'s own doc comment
(`internal/auth/store.go`) states the reason as written, that the leaf
name has one home rather than a literal at every call site; the
`ReadWritePaths` argument sits one comment above it, on `leafDirName`,
where it is attributed to R-5b11 and reads on the serve unit's grant
being one static path rather than one per record kind. The draft close
attributed that argument to a 5d cost; the code does not, and this
entry follows the code.

**Task 8** (5 named, `opus`): inserting through `atomicfile.WriteFile`,
checking existence before creating, sweeping on file mtime instead of
carried expiry, and deleting a malformed record on sight each map to
`TestInsertRefusesReplay`, `TestConcurrentInsertAdmitsOne`,
`TestSweepPrunesOnCarriedExpiry`, `TestRecordCarriesRetentionNotTokenExpiry`,
`TestSweepRetainsMalformedRecord`, and
`TestSweepRetainsCrashTruncatedRecord`. Dropping the directory `Sync`
call is **reported unfired by design**: crash consistency has no
gate-resident test. The argument does not live in `writeAndSync`'s doc
comment, which states only that the helper writes, syncs, and closes.
It lives in the `NonceStore` doc comment (the O_CREATE|O_EXCL
first-writer-wins paragraph, R11), in the inline comment above the
`syncDir` call in `Insert` ("a crash after a consumption can lose the
record while the link that produced it is still within its expiry"),
and in `syncDir`'s own doc comment (declared gap 2).

**Task 9** (6 named): the plan names six, and all six are listed here
rather than the draft close's five. Checking only the session's issue
time and not the token's mint time, keying the epoch on the raw
address, checking the epoch against the session's issue time rather
than its anchor, comparing at the epoch exclusively, and offering
publication without a read each map to `TestCheckLink`,
`TestCheckSessionNormalizesAddress`,
`TestCheckSessionRefreshedAfterBumpStillRefused`, the boundary row of
`TestCheckSession`, and `TestReadAuthoritativeWithoutSubscriber`. The
sixth, writing through `atomicfile.WriteFile` instead of `WriteStream`,
is **reported unfired by design**: the difference is the directory
fsync, whose failure needs a power cut; the `EpochStore` doc comment in
`internal/auth/epoch.go` carries the reason.

### The declared coverage gaps

**Six declared, against the running total of sixty-seven after rung 8c,
bringing the total to seventy-three.**

1. **A constant-time comparison has no gate-resident proof** (Task 5,
   mutation 4). `crypto/subtle.ConstantTimeCompare`'s use is provable
   only by reading the call; the criterion is a reviewer's read, not a
   test.
2. **Crash consistency rests on two fsync calls rather than on a test**
   (Tasks 8 and 9). A power cut is not a thing this gate reproduces;
   the `NonceStore` and `EpochStore` doc comments and `syncDir`'s carry
   the argument instead.
3. **No handler, no limiter, and no origin check exercises any of
   this** (R-5b1). 5b-ii is the first real caller of the token, the
   session, and both stores.
4. **The identity-source family has exactly one declared source**
   (`magic-link`). Its multi-source behavior
   (`TestIdentitySource_SharedKeyAcrossSources` and the collision-gate
   rows) is proved only against fixture registries; the second real
   source arrives with a forward-auth front nobody has built.
5. **Nothing validates a registered threshold's value except the two
   this pass's hook reaches** (Task 4, security review finding 24 from
   5a). Task 2's hook takes the environment snapshot, so
   `DUBPLATE_NONCE_RETENTION` and `DUBPLATE_MAGIC_LINK_EXPIRY` validate
   at load and their relation with each other (R31). Every other
   registered threshold this pass declares (`DUBPLATE_SESSION_LIFETIME`,
   `DUBPLATE_SESSION_ABSOLUTE_CAP`) still accepts any parseable value,
   which is the rung 1b wall contract's own property rather than a gap
   this pass created.
6. **The sweep's interval driver does not exist** (R-5b14). The sweep
   is proved by direct call rather than by elapsed time; no goroutine
   calls it on a schedule until the serve verb's lifecycle wires one at
   5b-ii.

### The review record, with its deviations

- **`code-simplifier` ran three batched per-segment rounds**
  (`8ae24cd`, `7d1f137`, `1d4ed29`) rather than one dispatch per task,
  a deviation from CLAUDE.md's stated cadence, recorded here rather
  than left silent.
- **The rung 5 security cycle did not run this window.** R-5b16's
  fresh-context `web-auth-security-reviewer` read was cut for budget.
  It is owed to the next attended session as that session's first item,
  before 5b-ii's own merge, and covering both halves per R-5b16.
- **The 5b-ii plan prep tail was cut** from the window for the same
  reason. 5b-ii's plan is authored just in time, after the security
  read.
- **Task 10's drafts escalated on seven draft-content findings.** The
  `diff-reviewer` read the close drafts against the task's acceptance
  criteria and escalated: a bare green with no measured evidence, a
  Task 9 row claiming six mutations while listing five, two doc-comment
  citations naming comments that do not carry the argument attributed
  to them (`writeAndSync` and `store.go`'s leaf-name reason), a
  three-round simplifier claim without its commits, a STATUS body over
  the sixty-line cap, and a dropped "awaiting Geoff's read" block. The
  conductor's fold corrected all seven against the tree; this entry and
  STATUS are the corrected text.
- **R-5b14's one-hour sweep period has no home in 5b-i.** The reviewer's
  note stands: an unused constant fails lint, so the constant lands
  with 5b-ii's interval driver, not here.
- Task 0 (`opus`, fresh context) and the in-chain `diff-reviewer` on
  every task are the review this pass ran in full.

### The spec errata this close owes

Drafted for the conductor to apply to
`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md` as dated
errata. None is applied yet; 5b-ii's plan prep is where they land.

1. **R23's ceiling table and the pass split** (R-5b1): 5b becomes 5b-i
   and 5b-ii, both ceilings and the rung total tracked against R23's
   15.75M reservation (set at 5a's own close).
2. **R24's review shape** (R-5b16): a fresh-context
   `web-auth-security-reviewer` read at each half's pass end, not one
   read "inside 5b" as R24's original wording states.
3. **R29's wording**: `Load` requires every owned-now wall row rather
   than naming `DUBPLATE_IDENTITY_SOURCE` by hand, closing risk 11 for
   every rung rather than one key.
4. **The env template's row assignment** (R-5b7): one literal row
   (`DUBPLATE_IDENTITY_SOURCE=magic-link`) lands at 5b-i rather than at
   5d, because `TestTemplateKeysCoverEveryOwnedNowRow` is a hard gate.
   `spec:2205`'s "commented pending rows only" and `spec:2170`'s row
   assignment both take the correction; the other six non-secret names
   stay 5d's.
5. **Two of `spec:2168`'s exit-gate assertions belong elsewhere**: the
   roster-removal half of the third assertion is 5b-ii's, and the
   verify-subrequest half of the eighth is 5d's, since `spec:2170`
   gives the endpoint to 5d.
6. **The staleness pre-task's `serve` reserved-class bullet**
   (`spec:2297-2300`) is stale; 5a removed that class.
7. **The `keys.go` pointers at `spec:1506`** (`:177-178` and
   `:171-176`) resolve to `:165-166` and `:159-162`.
8. **The `atomicfile` pointer at `spec:434`** (`:45-56`) resolves to
   `:55-91`, with the directory fsync at `:82-89`.
9. **The epoch's session check reads the anchor, not the issue time**
   (review B1). `spec:1085-1086` says "a session whose issue time
   predates the epoch is refused"; as landed the check reads the
   anchor, the field refresh never moves.
10. **The token's field list gains the mint time.** `spec:1016-1017`
    enumerates three fields where R30 (`spec:640-642`) requires a
    fourth; the landed token carries four.
11. **The serialization claim covers the wire form.** `spec:1017-1019`
    states the claim; the landed token states the encoding
    (base64url-then-join), so no field value can carry the separator.
12. **The nonce record's carried value.** `spec:1664-1668` says the
    record carries "its expiry" without saying whose; the record
    carries consumed-at plus the resolved retention.
13. **R38's carve-out covers the loopback set** (`spec:758-767`), not
    the literal name `localhost` alone.
14. **The non-secret name count.** `spec:1640-1645` says "eight" and
    enumerates seven; `spec:2170` gives 5d "its eight literal rows".
    One (`DUBPLATE_IDENTITY_SOURCE`) lands at 5b-i, so six remain for
    5d.
15. **R33 and R24 belong to 5b.** The 5a entry assigns R33 to this
    pass, and R-5b16 corrects R24's one-read-inside-5b wording.
16. **R1's own `keys.go` pointer at `spec:284`** (`:176-177`) moves to
    `:165-166`, the same move as item 7 above.
17. **The ceiling erratum's own pointer.** The plan cited
    `spec:2078-2087` for it in revision 1; the correct pointer is
    `spec:2042-2050`, and the rung reservation's live baseline is the
    15.75M that erratum set.
18. **R-5b14's own wording** describes the one-hour interval constant
    as landing at 5b-i; as executed, only the sweep-at-open call and
    the callable method landed, and the interval driver (constant
    included) is 5b-ii's in full, matching the resolution's operative
    clause.

### The 5b-i to 5b-ii handoff

5b-ii inherits: the identity leaf and its address normalizer, which
every keying, logging, and rate-limit lookup runs through
(`spec:1161-1163`); the token's verifier and the nonce it returns, to
be inserted after verification and never before; the token's version
label as the only place a later field, a destination included, may be
added (R33, `spec:690-692`); the session's minter, verifier, and
refresher, with the rule that refresh never moves the anchor, that the
anchor is the session's original issue time, and that refresh counts
same-origin requests only (`spec:1081-1082`); the epoch's two refusal
checks, the session check reading the anchor and never the issue time,
applied to a session on every request and to a token at consumption,
plus the roster re-check that refuses a link for an address removed
from the roster (R30, `spec:2168`), which 5b-ii still owes; the seam's
`Read` as the authority a credential store stamps from; the
state-directory opener and its naming rule, which the roster's own leaf
sits beside rather than inside; `registerAuthConfig` and the declared
key set, which 5b-ii's remaining thresholds join rather than re-derive;
the constraint that `internal/auth` never reads the environment; 5a's
own carries unchanged (`ClientIP` never re-derived, the origin check
inserts after the strip and after normalization, `Origin` and
`Sec-Fetch-Site` survive ingress, `CrossOriginProtection.Check` returns
nil when both are absent, every URL builds from the configured public
origin and never `r.Host`); the merge carry on
`internal/serve/server.go:161`, forwarded to 5d; the 30-second read
timeout and the `http.MaxBytesReader` alternative (R-V), until
`spec:1590`'s cap lands; and the two module dependencies R26 budgets,
both of which 5b-ii is the first pass to take, `go.mod` being unchanged
across this pass. **The nonce sweep's interval driver, a package
constant naming its one-hour period, is 5b-ii's to build alongside the
goroutine that calls `Sweep` on it** (R-5b14, corrected above). **The
fresh-context security review is owed before 5b-ii's own merge**,
covering both halves together per R-5b16.

### Both budgets

Tokens: this window's execution against the 5.0M ceiling (4.0M flag)
ran over, landing at roughly **5.5M** for Tasks 1 through 10 combined
with their fix and simplifier rounds. **The overage is recorded against
R23's live 15.75M reservation rather than funded from another pass**
(ruled, conductor 2026-09-12, applying arithmetic 4.25 + 5.0 + 4.5 +
4.5 + 3.0 = 21.25M against the live 15.75M). Attended time: **zero
planning misses** (Task 0 found no blocker the dispatch notes had not
already corrected) and **zero execution sittings** in this window; the
rung 5 security cycle and the plan-prep tail for 5b-ii were both cut
from the window for budget reasons rather than surfaced mid-task, which
is why they appear above as owed rather than as a sitting.

## Rung 8, pass 8c (absorb and repoint): closed 2026-09-12, lane 1

Tasks 0 through 4 executed against
`docs/superpowers/plans/2026-09-08-rung-8c-absorb-repoint.md` (revision
2) in the main tree, run `wf_c808ec5e-1ba`, with the close's own fix
rounds under `wf_5d503f3a-ebd`. Task 0 (`opus`, fresh context) found
no blockers against HEAD `6a9570d`: the lane registry, `flac.go`'s
three call sites, `buildFLAC`'s four external uses, `tagio`'s exported
surface, and every `spec:`/`docs/HISTORY.md:` pointer this plan carries
all resolved as expected. Baseline measurements: `internal/engine/shadow`
at 75 top-level tests, 140 RUN lines, 65 subtest passes, about 0.18s;
`TestConformanceLedgerSubjects` at twelve subtests, about 0.05s.

### The plan's own prep, carried from STATUS

The plan was drafted, reviewed, folded to revision 2, and committed at
`5377d9f` against a 2.1M ceiling (the spec proposed 1M) with a 1.68M
flag. Review record
`research/2026-09-08-rung-8c-plan-review.md`: 13 findings folded, none
rejected, three conductor-owed folds (R1 as a stated departure from
R12's rung 9 assignment; the `Parse` read path; R4's composed
denominator). The prep workflow was `wf_0663b27f-52a` (seven agents,
1.99M), its cache session-bound. The conductor-owed folds across both
lanes' plans were ruled by recommendation on Geoff's "then start"
(2026-09-12 00:44), the ratify-or-reverse table in
`plans/2026-09-12-overnight-window.md`.

### What landed

Task 1 (`9be7e43`) moved every `internal/engine/shadow` test fixture
onto the writer: a new `fixture_test.go` carries the committed 42-byte
oracle template as a seed literal (moved out of `flac_test.go` with its
provenance comment) and a one-second variant (a two-byte splice of the
zero-sample template's STREAMINFO pack, encoding 44100 total samples at
the template's own 44100 Hz rate), and one `writeFixtureTrack` helper
writes the seed, calls `tagio.Write` with a `tagio.TagSet` and a nil
picture set, and pins the modification time after the write. Every
fixture-building call site in `capture_test.go`, `run_test.go`, and
`observe_test.go` converts from `"KEY=value"` string lists to
`tagio.TagSet`. **Task 1 took the plan's first offered route for the
content-join criterion**: every fixture now seeds from the one-second
variant, so `reconcile.go:242`'s title-and-length join compares a live
1.0-second value on both sides rather than vacuously comparing zero to
zero. `buildFLAC` still compiled, with every caller confined to its own
file, clearing the way for Task 2's deletion. All five named mutations
ran and restored; none was reported unfired for an unexpected reason
(mutation table below).

Task 2 (`3a299ac`, fix `6de7d55`, conductor-owed R6 fix `12d1e97`)
deleted `internal/engine/shadow/flac.go` and `flac_test.go` (313 and
361 lines: the block walker, the Vorbis comment decode, `safe_cast`,
date splitting, `buildFLAC`, `blockHeader`, `writeTemp`, and
`TestBuildFLACEmitsTheStreaminfoLayoutByteForByte`), and created
`libraryread.go` + `libraryread_test.go` and `itembag_test.go`.
`readFLAC` keeps its name and signature (R7/A3): `run.go:209` and
`capture.go:158` needed no edit. It now calls `tagio.Parse` once per
file and the newly exported `tagio.Tags` for the tag set (R1 as
applied, the 8b handoff's second half), reading `Container.StreamInfo`
for the length with A4's zero-rate guard preserved. The eighteen-row
`matcher.ItemFields` projection survives with its per-row mediafile
citations (restored by the fix round after the main commit dropped
them) and a new sentence naming ruling 3 as the reason the projection
stays in `shadow` rather than moving to `tagio` with the rest of the
container knowledge. Under R1, `tagio` gained one exported symbol,
`Tags(c Container) TagSet` (`read.go:39-45`), documented in `doc.go`
as `Read`'s own unexported `project` exported unchanged; `Read`'s
behavior is byte-identical.

R2's seven-row destination table (as executed, below) was discharged
test by test, including three previously-ungraded `TestParseRefusals`
rows this task was pre-authorized to add ("STREAMINFO is not the first
block" in both its trigger shapes, "no STREAMINFO block", and "carries
no '='") and one previously-uncovered barcode-precedence row folded
into `tagio`'s own `TestProjectLastStyleRule`. **The conductor-owed R6
fix** carries the invalid-UTF-8 value change through `shadow`'s own
projection test: `spliceInvalidUTF8TitleByte` (`fixture_test.go`)
length-preserves a byte splice of a writer-built fixture's title byte
to an ill-formed UTF-8 continuation byte, because `tagio.Write` only
ever accepts values already valid UTF-8 (ruling 10) and `shadow` has no
lower-level fixture path; `TestReadFLACProjectsInvalidUTF8AsReplacementChar`
asserts `readFLAC` projects U+FFFD, proving `shadow`'s reader carries
`tagio`'s own rule through rather than duplicating it. All seven named
mutations ran and restored, including a new test
(`TestReadFLACProjectsAbsentFieldsToTypedZeroValues`) the implementer
added because the plan's own all-fields-present fixture test could not
catch mutation 5 (an unguarded absent-field dereference) on its own.

The batched `code-simplifier` round (`45dd73a`) derived the one-second
seed by `slices.Clone` plus a two-byte patch instead of a second
40-byte literal that could silently drift from the first (the two
literals' doc comment had already gone wrong once, naming three bytes
for a two-byte change), collapsed `stringVal`/`intVal`/`boolVal` into
one generic `deref[T]` (three copies of one sentence about
`ProjectItemFields`' typed-default rule), and rebuilt the
second-STREAMINFO refusal fixture with `slices.Insert` rather than a
nested append over an aliased prefix. No behavior or assertion change;
the derived seed is verified byte for byte against both the prior
literal and `internal/harness/oracle/testdata/template.flac`.

Task 3 (`67d5c3d`) added the pass's one behavioral change: a
capture-path test proving a container the deleted reader accepted but
`tagio` refuses (ruling 9's "more than one VORBIS_COMMENT block") is
**discarded whole**, matching the partial-read rule `captureAlbum`
already applies to a truncated file (zero captures, zero exclusion
rows, one partial-read row). A second test proves the same album
unspliced still captures normally, so the witness fixture is a witness
rather than a broken file. The witness's bytes come from the writer's
own output plus `spliceDuplicateVorbisComment`, which parses the
writer's output via `tagio.Parse`, identifies the VORBIS_COMMENT block
by matching each block's payload against
`Container.VorbisComment.Vendor`'s own bytes (no length-prefix decode
needed, since STREAMINFO is a fixed 34-byte layout and PADDING is
zero-filled), and copies that block's header-plus-payload whole,
splicing a byte-identical second copy after the first: no block-type
constant, no header encoder, no `encoding/binary` import anywhere in
`shadow` (A1). All three code/test mutations ran and restored; the
fourth named mutation (enumerate from the deleted reader's behavior
rather than `tagio`'s refusal set) is a documentation-construction
check rather than an executable one, verified by construction instead.
**The enumeration table went through two comment-only review rounds**
before it was correct (below).

### The A1 through A9 coverage table

| Item | Go site | What proves it |
| --- | --- | --- |
| A1 (no container knowledge in `shadow`'s non-test files) | `internal/engine/shadow/libraryread.go` (whole file) | `grep -rn "encoding/binary\|fLaC\|VORBIS" internal/engine/shadow` returns nothing outside a comment and outside `capture_test.go`'s pre-existing `[]byte("fLaC")` literal in `corrupt()` (rung 4a Task 6, untouched by this pass, out of A1's non-test-file scope). Task 3's splice helper touches container bytes without a block-type constant or header encoder, exactly as A1's scoping sentence permits. |
| A2 (eighteen-row projection, per-row mediafile citations, typed-default rule) | `projectItemFields` (`libraryread.go:45-`) | `TestReadFLACMapsEveryVorbisKey` grades seventeen distinguishable rows plus `DataSource`'s asserted empty string; the citations were restored by Task 2's fix commit `6de7d55` and verified byte for byte against the deleted `flac.go`. |
| A3/R7 (`readFLAC` keeps its name and signature) | `run.go:209`, `capture.go:158` | Both call sites are untouched in the diff; `git diff` across the whole pass shows zero lines changed in either file's call site. |
| A4 (length division, zero-rate guard) | `readFLAC` -> `trackLength` path in `libraryread.go` | `TestReadFLACLengthFromStreaminfo`'s table: ten seconds at 44100 Hz, a fractional length, and a zero sample rate reading zero rather than dividing by zero. Named mutation 2 (drop the guard) fires this row. |
| A5/R1 (one parse per file) | `readFLAC` (`libraryread.go:24-29`) | One `tagio.Parse` call, feeding both `tagio.Tags` and `Container.StreamInfo`; the doc comment states the count. |
| A6 (provisional framing rewritten) | `libraryread.go`'s package/function doc comments | `flac.go:62-70`'s provisional paragraph and rung-4-calendar rationale do not survive; the deleted file's whole text is gone with it. |
| A7/ruling 4 (`buildFLAC` does not survive) | deleted `flac_test.go` | `grep -rn buildFLAC internal/engine/shadow` returns nothing after Task 2's commit. |
| A8 (tightening asserted and counted) | `capture_test.go`'s two new tests | `TestCaptureDiscardsATrackWithADoubledVorbisCommentBlock` and its unspliced companion, plus the eight-row enumeration below. |
| A9 (conformance instrument stays green) | `internal/engine/tagio/conform_test.go:640` | `go test ./internal/engine/tagio/... -run Conformance -count=1`: `TestConformanceLedgerSubjects`, 12 subtests, about 0.06-0.07s, all green, reconfirmed at this close. |

### The seven Resolutions, retired

**R1** (the shadow parses once; `tagio` exports its tag projection).
**Conductor-owed. Ruled 2026-09-12, as applied**: the shadow consumes
`Parse` and the newly exported `tagio.Tags`, taking the 8b handoff's
second half rather than its first (`docs/HISTORY.md`'s 8b entry,
"registry consumer" sentence); `tagio` gains one exported symbol;
8b's declared coverage gap 4 retires now rather than at rung 9.
Site: `internal/engine/tagio/read.go:39-45` (`Tags`),
`internal/engine/shadow/libraryread.go:24-29` (`readFLAC`). **Had R1
been ruled the other way** (the fold), the shadow would call
`tagio.Read` for the tag set and `tagio.Parse` for the STREAMINFO, two
parses and three file opens per track, `tagio` would gain no exported
symbol, and 8b's coverage gap 4 would stay open for rung 9 to close
under R12's original assignment.

**R2** (the five reader tests get named destinations, two staying in
`shadow`). **Conductor-owed. Ruled 2026-09-12, as applied**: the
seven-row destination table stands (below). **Had it been ruled
otherwise**, the two staying tests would be rewritten inside `tagio`
against `TagSet` expectations, and the eighteen-row `ItemFields`
projection would carry no test of its own.

**R3** (fixtures come from the writer; the one malformed fixture is a
byte splice). Applied. Site: `internal/engine/shadow/fixture_test.go`
(`writeFixtureTrack`, `spliceDuplicateVorbisComment`,
`spliceInvalidUTF8TitleByte`). **Ruled otherwise**, the malformed
fixture would be a hand-written byte literal transcribed in full.

**R4** (the exclusion count: exclusion set plus a read-only denominator
plus a bounded count). Applied for items 1 and 3 (Task 3's enumeration;
the count stated as zero known exclusions out of the denominator, bound
by the exclusion set, next measurable at the shadow's next deploy).
**Item 2, the denominator's composition, is conductor-owed. Ruled
2026-09-12**: `ObservedApplied` + `ObservedSkipped` + the
`Unreconciled` counts, read by the conductor through `dubplate-server
shadow report`, never the bare binary as root. **The read itself did
not happen this pass**: the conductor's box read failed at 03:14 AKDT
(ssh timed out at the connection level, twice; no quiet-box check was
possible). **The R4 denominator is therefore unread, because the box
was unreachable at 03:14 AKDT, and the read is owed to the next
attended session** before rung 15 treats R4 as final. Had R4 item 2
been ruled otherwise, the conductor would take a read-only sample of
library and inbox files off the box and grade them on the workstation
with the landed parser.

**R5** (the `io.ReaderAt` gap declined at 8c, handed to rung 9).
Applied; no task in this pass widened `Parse`'s signature. The gap
stays declared (restatement of 8a's gap 7, not counted again), now
naming rung 9 as the decision-maker rather than "8c or rung 9." Had it
been ruled otherwise, `Parse` would gain a second, injectable
`io.ReaderAt` entry point and a mid-chain I/O error test.

**R6** (the invalid-UTF-8 value change stated in advance). Applied,
with the conductor-owed extension landed at `12d1e97`: R6's finding
during Task 2's second review round (no row in `shadow` graded the
value change, only `tagio`'s own `TestRulesInvalidUTF8ReadBack`) forced
a choice between giving `shadow` a low-level fixture path, exporting a
`tagio` test helper, or ruling R6 already discharged in `tagio`. The
conductor ruled a fourth option not named in the plan's own text: a
length-preserving byte splice over writer output, the same shape R3
already establishes for the malformed-container fixture, carrying the
value change through `shadow`'s own projection test without exporting
anything new from `tagio`. This is a **named deviation from the plan's
own three-way framing of R6's escalation**, resolved with R3's own
method rather than any of the three options the escalation posed.

**R7** (`readFLAC` keeps its name and signature). Applied; `run.go:209`
and `capture.go:158` needed no edit, confirmed by the implementer's
grep and by this close's own re-check. Had it been ruled otherwise,
the function would be renamed and both call sites plus
`reconcile.go:35` would need editing together.

### R2's destination table, as executed

| Test | Destination |
| --- | --- |
| `TestReadFLACMapsEveryVorbisKey` | Stayed in `shadow`, rewritten as `libraryread_test.go`'s eighteen-row table over a writer-built fixture. |
| `TestReadFLACStylePrecedence` | Moved. Its sixteen rows were walked against `tagio`'s existing projection tests: fifteen already covered by an existing row over the same branch; one uncovered row (barcode falling through the four read-only aliases in order) added to `TestProjectLastStyleRule` (`internal/engine/tagio/read_test.go`). |
| `TestReadFLACLengthFromStreaminfo` | Stayed in `shadow`, rewritten as a table over `tagio.StreamInfo` values needing no file. |
| `TestReadFLACParsesTheCommittedTemplateBytes` | Retired. Its purpose is already covered by `tagio`'s `TestParseGolden` (`container_test.go:359`); no second external pin added, and the now-dead `committedTemplateFLAC` literal was removed with it. |
| `TestReadFLACRefusesAMalformedContainer` | Split. The container half moved to `tagio`'s `TestParseRefusals`, which also gained three previously-ungraded rows this task was pre-authorized to add; the capture half stayed in `shadow` as a thin test asserting `readFLAC` returns the zero `ItemFields` and an error, never a partly filled bag. |
| `TestItemBagRoundTripsThroughProjectItemFields` | Moved within `shadow`, to the new `itembag_test.go`; `itembag.go:11-13`'s pointer repointed. |
| `TestBuildFLACEmitsTheStreaminfoLayoutByteForByte` | Deleted with `buildFLAC` (ruling 4). |

### Task 3's enumeration, verbatim as the tightening record

The table below is the corrected, final form: Task 3's own report
carried seven rows; a review round added row 2 (a shape the deleted
reader accepted and the report's first draft omitted: more than one
STREAMINFO block); a second review round found row 8 (as it then
stood) named the wrong `tagio` test. Both corrections are folded here
because no further fix round ran to fold them into the task's own
committed return, and re-verified against the tree at this close
(`go test ./internal/engine/tagio/ -run 'TestParseRefusals|TestParsePictureRefusals' -v`).

| # | Shape | Landed reader's old behavior | `tagio` test | Reachable from a real upload? |
| --- | --- | --- | --- | --- |
| 1 | More than one VORBIS_COMMENT block | Read happily, took the first value per key | `TestParseRefusals/second_VORBIS_COMMENT_block` (asserted directly by Task 3's capture-path test) | Crafted file only; no known compliant encoder emits this. |
| 2 | More than one STREAMINFO block | Parsed each occurrence and overwrote `file.info` with the last, no error | `TestParseRefusals/a_second_STREAMINFO_block` | Crafted file only, no known compliant encoder emits a second STREAMINFO. |
| 3 | A STREAMINFO block that is not first, or no STREAMINFO block at all | Neither read as an error; projected a zero length | `TestParseRefusals/first_block_is_not_STREAMINFO` and `TestParseRefusals/no_STREAMINFO_block_at_all` | Crafted file only. |
| 4 | A block appearing after a header whose last-block flag is set | Stopped at the flag, returned partial | `TestParseRefusals/block_after_a_header_whose_last-block_flag_is_set` | Crafted file only. |
| 5 | More than one SEEKTABLE block | Skipped extras by declared length | `TestParseRefusals/second_SEEKTABLE_block` | Crafted file only. |
| 6 | A comment field name carrying a code point outside U+0020-U+007E | Upper-cased it and kept it | `TestParseRefusals/comment_field_name_above_U+007E` | Unlikely from a compliant encoder; possible from a hand-edited or non-compliant tagger. **Finding, execution-discovered**: `validateFieldName` (`container.go:347-350`) refuses both directions (`r < 0x20 \|\| r > 0x7e`), but `TestParseRefusals` has no dedicated subtest for the below-U+0020 direction. This is a coverage gap in `tagio`'s own test suite, not a missing enumeration row (the shape ties to a real, if partial, test). |
| 7 | A comment carrying no `=` | Skipped that comment, read the rest | `TestParseRefusals/comment_carries_no_'='` | Possible from a non-compliant real-world tagger; more plausible than rows 1-5. |
| 8 | A PICTURE block whose declared lengths do not fit its payload | Skipped every PICTURE block by declared length | `TestParsePictureRefusals` (`internal/engine/tagio/picture_test.go:91`; five subtests: `media_type_length_overdeclares`, `description_length_overdeclares`, `data_length_overdeclares_near_the_block_header's_own_3-byte_maximum`, `block_shorter_than_the_32-byte_minimum`, `media_type_consumes_the_block,_no_room_left_for_geometry`) | Plausible from a real corrupted or truncated upload, not only a crafted one. |

Every row ties to a named, currently-passing `tagio` test; none is a
row with no test to name.

### R4's three-part answer to the exit gate's count

1. **The exclusion set**: the eight-row table above, each row a
   workstation measurement over committed code, needing no box.
2. **The denominator**: **UNREAD**. Conductor-owed R4 item 2 ruled the
   composition (`ObservedApplied` + `ObservedSkipped` + `Unreconciled`),
   but the read itself failed: ssh to the box (5.78.148.42) timed out
   at the connection level, twice, at 03:14 AKDT, so no quiet-box check
   was possible and `dubplate-server shadow report` was never read.
   Owed to the next attended session, before rung 15 treats R4 as
   final.
3. **The count**: zero known exclusions out of the (unread)
   denominator, bounded by the exclusion set above. Not derivable from
   stored records, because a stored capture carries no container bytes
   (`capture.go:143-162`). The tightening reaches the live capture at
   the shadow's next deploy, not at this merge.

### The declared coverage gaps

Four new gaps this pass creates, joined to the running total of
sixty-three (rung 3's twenty-one, rung 4a's nine, rung 4b's four, rung
1b's two, rung 8a's seven, rung 8b's nine, rung 5a's eleven), bringing
the total to **sixty-seven**. Two items this pass's own expected list
carried are restatements of an already-counted gap rather than new
declarations and are not counted again: "the `io.ReaderAt` gap"
restates 8a's gap 7 (now naming rung 9 alone as the decision-maker),
and "no real library file has been read through the absorbed path"
restates 8a's gap 6.

**A fifth item from the plan's expected list, "the end-to-end length
join's length term is vacuous," does not exist**: Task 1 landed the
non-zero-sample seed variant, so every fixture's join compares a live
1.0-second value rather than zero to zero, exactly the condition the
plan named as retiring this gap.

1. **The tightening's live effect is unmeasured** (R4). The count is
   bounded rather than counted, because stored captures carry no
   container bytes.
2. **The invalid-UTF-8 value change is graded on a synthetic fixture**
   (R6), not on a file any member uploaded, and the fixture reaching
   `shadow`'s own projection test is itself a byte splice rather than a
   file any encoder produced.
3. **The content join's length term discriminates a real length from
   zero or a missing one, but not between two different non-zero
   lengths**, because Task 1 seeded every fixture in the package from
   the same one-second variant. This is the corrected, narrower shape
   of the plan's originally-feared gap 5: the vacuous zero-to-zero
   comparison does not exist, but the join still cannot distinguish
   fixture A's length from fixture B's, since both are 1.0 seconds by
   construction. Flagged by Task 1's own `diff-reviewer` round as worth
   carrying here.
4. **`TestParseRefusals` has no dedicated subtest for a comment field
   name below U+0020**, though `validateFieldName` refuses both
   directions. Discovered while building Task 3's enumeration (row 6,
   above); the refusal code path is exercised, the specific below-range
   shape is not independently graded.

### 8b's declared coverage gap 4 is retired

R1 stood as applied: the shadow's `readFLAC` now reads a file's tags
and its STREAMINFO from one `tagio.Parse`, so a caller wanting both no
longer parses twice. 8b's gap 4 ("`Read` returns no pictures, so a
caller wanting the tags and the cover parses the file twice") does not
apply to `shadow`'s own reader, whose caller now gets the tag set and
the length from one parse; the gap's remaining scope is a caller
wanting `Read`'s tag set alongside `Container.Pictures`, which is
`Read`'s own signature and unaffected by this pass. **The running gap
total above is not reduced for this retirement**: the sixty-three
figure and the sixty-seven figure both count declarations, not a live
open list, so retiring one item changes what rung 15 should read as
still open, not the historical count.

### The mutation table

**Task 1** (5 named, all fired or confirmed unfired for the stated
reason): (1) skip the `tagio.Write` call, nine tests across
capture/observe/run fail because tracks carry no tags; (2) pin the
mtime before the write, `TestRunIsDeterministicUnderAFrozenClock` fails
because the atomic rename installs a fresh mtime; (3) give both
`windroseTracks` entries the same title, no test fails, confirmed
unfired for the stated reason (the capture groups by
albumartist/album, invisible at this altitude); (4) fixture mode 0
instead of 0o644, ten tests fail because the file cannot be reopened;
(5) a non-nil empty picture set instead of nil, no test fails,
confirmed unfired for the stated reason (the seed carries no PICTURE
block, so strip and leave-alone are byte-identical). Every restoration
verified with `git diff` empty.

**Task 2** (7 named, all fired or confirmed for the stated reason):
(1) skip the STREAMINFO read, `TestReadFLACMapsEveryVorbisKey`'s
Length row fails (1→0); (2) drop the zero-rate guard, the zero-rate
row fails with `+Inf`; (3) swap `Label` for `Country`, both rows fail,
each named; (4) keep the deleted `parseFLAC` and call it instead of
`tagio`, the package still compiles and every landed `shadow` test
passes, confirming Task 3's tightening test is what catches this, not
Task 2's own criteria; (5) an unguarded `*string` dereference, panics,
caught only after the implementer added
`TestReadFLACProjectsAbsentFieldsToTypedZeroValues` because the plan's
own all-fields-present fixture could not exercise an absent field; (6)
return the first matching comment's value versus the last for a
repeated key, confirmed unfired in `shadow`, for the stated reason (the
rule is now `tagio`'s, graded at `read_test.go`); (7) leave
`flac_test.go` in place while deleting `flac.go`, the package fails to
compile, the mechanical proof the reader and its tests die together.
Every restoration verified with `git diff` empty.

**Task 3** (4 named): (1) build the witness fixture without the
splice, `captures = 1, want 0`, the album was captured whole; (2) catch
the read error in `captureAlbum` and continue, `captures = 1, want 0`,
the bad track silently dropped, exactly the tolerance this pass
removes; (3) assert the album is excluded with a reason rather than
discarded, fails against `capture.go`'s own discard-not-exclusion rule
(`excluded = [], want one row`; `partial reads = [...], want none`);
(4) enumerate the tightening from the deleted reader's behavior rather
than `tagio`'s refusal set, not an executable mutation, verified by
construction instead (every row built from `TestParseRefusals` and
`TestParsePictureRefusals` subtest names directly; the one gap found,
row 6's untested below-U+0020 direction, is reported above rather than
silently dropped or invented). Every code/test mutation's restoration
verified with `git diff` empty.

### The review record, with its deviations

Task 1 was accepted first pass. Task 2's R6 row was accepted on the
implementer's report without a `diff-reviewer` read, a budget
deviation recorded here rather than left implicit. Task 3 and Task 4
each carried a second fix verdict resolved by the conductor: Task 3's
was the enumeration row naming (row 8's wrong `tagio` test), and Task
4's was six pointer digits corrected across the close drafts and the
spec erratum (`5d7c42a`, then `d2914a2`).

### What rung 9 inherits

`tagio`'s exported surface after this pass: `Parse`, `Read`, `Write`,
`Tags` (new this pass), `TagSet`, `Mapping()`, `WriteOrder()`,
`Serialize`, `EmptyString`, `EmptyList`, `Container`, `Block`,
`StreamInfo`, `Comment`, `VorbisComment`, `KeyWrite`, `Field` and its
69 constants, `FieldMapping`, `Style`, `Picture`, `FrontCover`,
`Format`, `SupportedFormats`, `Sniff`, `ErrUnsupportedContainer`,
`ErrMalformedContainer`, `Deviations`. The `io.ReaderAt` call (R5) is
rung 9's alone to make: whether `Parse` grows a second, injectable
entry point is undecided, and no consumer in this pass forces it. The
combined-read question R12 posed is **answered, not left**: R1 as
applied means a caller wanting one file's tags and its STREAMINFO
parses once through `Parse` plus `Tags`, the shape `shadow` now uses;
rung 9 inherits that shape as precedent rather than a decision still to
make, though `Read`'s own tag-only contract (R12, 8b) is unchanged and
a caller wanting `Read`'s tag set alongside `Container.Pictures` still
parses twice. The tightening's live effect on rung 7's intake gate
bound does not reach the box until the shadow's next deploy; this
pass's merge changes no running process.

### The spec corrections

Applied directly to
`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md` as a dated
erratum, in the shape the 8a-close and 8b-close errata used. See the
diff for the exact inserted text; summarized here:

1. **The five-tests sentence** (spec's pass-split table's 8c row,
   `spec:1226`) is corrected to R2's seven-row destination table,
   naming why two of the five stay in `shadow` (their subject is
   `matcher.ItemFields`, and ruling 3 forbids `tagio` naming that
   type).
2. **The reference-count sentence** ("What the absorption costs,"
   `spec:1036`) is corrected: the paragraph named six references
   (three call sites, two prose mentions, the reader's own doc
   comment). Under R7 two of the three call sites and both prose
   mentions need no edit, because `readFLAC` keeps its name.
   **`itembag.go:11-13` is a seventh reference the spec never named**,
   and it is the one this pass actually edits (its pointer at the
   deleted `flac_test.go` repointed to `itembag_test.go`).
3. **The ceiling table's 8c line** (`spec:1272`, "Propose 4M for 8a, 3M
   for 8b, and 1M for 8c," and the 8b-close erratum's restatement at
   `spec:85`) is corrected from 1M to this plan's 2.1M, re-derived from
   8b's actual under the spec's own re-derivation clause. Rung 8's
   ceilings now stand at **9.85M** against the spec's 8M reservation:
   4.5M (8a) plus 3.25M (8b) plus 2.1M (8c). 8c's actual spend is
   stated in this entry's own spend section.
4. **The read-path strictness section** ("The read path's strictness
   is a deliberate change, and 8c owns it," `spec:1070`) is corrected
   against Task 3's enumeration: the spec's own prose named the
   tightening only in general terms ("a real inbox FLAC with two
   VORBIS_COMMENT blocks or an out-of-order STREAMINFO"); the landed
   refusal set is the full eight-row table above, materially wider
   (it also refuses a doubled STREAMINFO, an out-of-range SEEKTABLE
   count, an out-of-range comment field name, a comment carrying no
   `=`, and a malformed PICTURE block), and the correction names the
   table as the authoritative enumeration rather than the spec's own
   two-example prose.
5. **CLAUDE.md's gate sentence** was checked: this pass adds no
   tracked Python file, so the sentence's Python wording stands
   unchanged.

### The rung 8 close statement

Rung 8 is closed: three passes (8a the writer core and reader, 8b
PICTURE and the conformance instrument, 8c the absorption and
repoint), `internal/engine/tagio` born and now the sole holder of FLAC
container knowledge in the repo, `internal/engine/shadow` repointed to
it with the tightening asserted rather than discovered.

**The five owner rulings** (`spec:209-247`):

1. FLAC only at rung 8, the per-format seam shaped for later passes.
   Discharged at 8a's close, the seam ruling 14 describes landed there.
2. The wazero-taglib spike retired, F1 superseded. Discharged at 8a's
   close.
3. The writer writes and strips the PICTURE front cover at launch.
   Discharged at 8b's close.
4. ID3v2.4 recorded as the future MP3 version, no ID3 written at rung
   8. Carried; rung 8 writes no ID3, so nothing exercises the ruling
   yet, and it stays recorded for whichever pass lands MP3 writing.
5. Fork B, the format ladder. Discharged: Geoff ruled 2026-09-08,
   FLAC-only 1.0, format ladder at 1.1 (`ROADMAP:formats-1-1`).

**The fourteen conductor rulings** (`spec:303-400`), one status each:

1. The sixteen ledger-silent keys and every corpus-unreachable case.
   Discharged at 8a's close.
2. An applied-field disagreement routes into
   `corpus/analysis/adjudications.json`. Discharged at 8a's close.
3. `tagio` does not depend on `matcher`. Discharged this pass: the
   eighteen-row `ItemFields` projection stays in `shadow`, and the same
   ruling is the stated reason Task 2's projection test does not move
   to `tagio`.
4. `buildFLAC` does not survive. Discharged this pass: deleted with
   `flac.go` in Task 2.
5. Full rewrite only at launch, one PADDING block per write.
   Discharged at 8a's close.
6. 8 KiB PADDING reservation. Discharged at 8a's close.
7. Unknown blocks preserved unconditionally. Discharged at 8a's close.
8. The frame region is the content hash's extent. Carried; no rung 8
   task computes a content hash, so the ruling stays a forward hook for
   rung 6 and rung 7.
9. The writer refuses a container it did not fully understand.
   Discharged at 8a's close for the writer, and discharged again this
   pass for the reader: Task 3 asserts `shadow`'s capture applies the
   same refusal set.
10. Field names and values follow RFC 9639 plus beets; invalid UTF-8
    replaced with U+FFFD on read. Discharged at 8a's close, and this
    pass's R6 fix (`spliceInvalidUTF8TitleByte`,
    `TestReadFLACProjectsInvalidUTF8AsReplacementChar`) carries the
    same rule through `shadow`'s own projection test.
11. `internal/atomicfile` grows a streaming write in pass 8a.
    Discharged at 8a's close.
12. The streaming write fsyncs the containing directory. Discharged at
    8a's close.
13. Rung 8 declares no configuration key. Discharged at 8a's close;
    this pass declares none either.
14. The seam is a registry keyed by container, one entry, no
    interface. Discharged at 8a's close; this pass's R1 keeps
    `shadow`'s read path off that registry by deliberate choice, not by
    reopening the ruling.

**Hooks 13, 14, and 3**: hooks 13 and 14 landed at 8a (the writer core
and the 69-key value model, per 8a's own close entry). Hook 3 landed
at 8b (the PICTURE decode and `Container.Pictures`, per 8b's own close
entry). **8c lands no hook**; under R1 it improves the shape hook 3's
consumer (rung 9's admit path) will use, because a file's tags and its
embedded picture now reach a caller off fewer parses, but that is a
consequence recorded here, not a hook this pass claims.

**The deviation note, as a statement rung 15 can read without
interpreting a file**: 8c's read path never reaches `tagio`'s seam
registry (R1's second deviation, folded with R1's own ruling); the two
reader tests that stayed in `shadow` are a stated departure from
`spec:1011`'s original five-tests sentence (R2), corrected above; the
2.1M ceiling is a stated departure from the spec's proposed 1M,
corrected above; `tagio` gaining one exported symbol at a pass the
spec called mechanical is a stated departure from a decision 8b closed
(R1, conductor-owed); and R6's escalation resolved with a fourth
option (a length-preserving byte splice) the plan's own three-way
framing did not name, recorded above under R6's retirement.

**Fork B recorded as ruled**: unchanged by this pass; FLAC-only 1.0,
the format ladder at 1.1 (`ROADMAP:formats-1-1`), per Geoff's
2026-09-08 ruling.

**The one carry that leaves the rung open**: the group non-FLAC
question (`docs/HISTORY.md`'s rung 8a entry) stays carried and is not
decided by this pass, which absorbs and repoints the FLAC reader alone.

### The evidence

`internal/engine/shadow`'s suite: 75 top-level tests / 140 RUN lines /
65 subtest passes / about 0.18s before this pass (Task 0's baseline);
76 top-level tests / 120 RUN lines / 44 subtest passes / about 0.2-0.5s
across runs after (Task 2 net -2, Task 3 net +2, the R6 fix net +1:
75-2+2+1=76). `internal/engine/tagio`'s conformance run:
`TestConformanceLedgerSubjects`, 12 subtests, about 0.06-0.07s, before
and after unchanged. `go test ./... -count=1` (every package): green,
2m19.516s. `bash scripts/check.sh`: green, about 2m43s, gitleaks the
dominant cost at 2m27s scanning about 119.64 MB.

Line counts across the pass (Task 1 through the `code-simplifier`
round, `6a9570d..45dd73a`, `internal/engine/shadow` and
`internal/engine/tagio` combined): **616 insertions, 738 deletions**,
across 14 files, a net deletion of 122 lines. `flac.go` (313 lines) and
`flac_test.go` (361 lines), 674 lines of container-decode code and its
tests, are gone; `libraryread.go` (90), `libraryread_test.go` (169),
`itembag_test.go` (40), and `fixture_test.go` (151) replace them
alongside small `tagio` additions (`doc.go` +10, `read.go` +12/-,
`container_test.go` +35, `read_test.go` +1).

The mutation table above: every named mutation across Tasks 1 through
3 fired or was confirmed unfired for its stated reason; none fired for
an unexpected reason and none was silently dropped.

### Spend

Budget: 2.1M ceiling. This close does not have exact per-agent token
figures on hand; the run's journals (`wf_c808ec5e-1ba`, ten
implementer/reviewer agent turns across Tasks 1-3, plus Task 0's
separate fresh-context `opus` dispatch and Task 4; then
`wf_5d503f3a-ebd` for the close's fix rounds) are the source if an
exact number is wanted. Attended interactions: **zero planning
misses**; **one execution sitting** is arguable, since the conductor's
R6 ruling and the box-read attempt both happened without a Geoff
interaction (conductor-owned, not an attended sitting), so the target
of zero execution sittings holds as reported, pending the conductor's
own accounting of turns it spent on R1, R2, R4, and R6.

### What a later pass would be wrong to rediscover

`shadow`'s read path consumes `tagio.Parse` and `tagio.Tags` directly
and never reaches the seam registry; a future pass adding a second
format to the registry does not make `shadow`'s FLAC path a registry
consumer by that addition alone, because the decision not to route
through it is R1's own, not an oversight. `readFLAC`'s name and
signature are a deliberate no-op (R7): a reader expecting the function
to be renamed alongside the container-knowledge move is reading the
plan's own two-name collision (`shadow.readFLAC` and `tagio.readFLAC`,
different packages, never confused at compile time) rather than a
defect. The content-join's length term now discriminates a real length
from zero, but every fixture in the package shares one seed value, so
a future test relying on the join to distinguish two different
non-zero lengths needs its own distinct seed, not this pass's shared
one (declared gap 3, above). The invalid-UTF-8 fixture path
(`spliceInvalidUTF8TitleByte`) is `shadow`'s only way to put an
ill-formed byte in front of a reader, because `tagio.Write` enforces
valid UTF-8 on write (ruling 10); a future `shadow` test wanting
another invalid-byte shape reaches for the same splice technique
rather than a new writer bypass. R4's denominator is unread, not
zero-valued: a later pass reading `docs/STATUS.md`'s carried-gap count
should not treat "zero known exclusions" as "zero exclusions
happened."

## Configuration homes spec: design sitting, 2026-09-11 to 2026-09-12

Not a pass. One brainstorm sitting with Geoff produced
`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`
(revision 1 at `5e4feef`, revision 3 at `07f4e01`), settling the four
homes (env wall, store settings, member preferences, machine file),
collections carrying writer, owner, and visibility, and the four attach
verbs. Five fresh-context `claude-opus-5` reviews shaped revisions 2
and 3 (records under `research/2026-09-11-configuration-homes-review-*`),
and Geoff ruled eight points, all dated in the spec. What a later pass
would be wrong to rediscover: the layout template is a filesystem-write
primitive and must be a closed grammar yielding path components; a
mirror push is untrusted input the server recomputes; identity is an
audio key computed by decode (encoders zero the header MD5) with a
separate file hash for drift; hardlink sharing needs the break-link
rule because beets writes in place; and settings read by other systemd
units cannot live in the store. Two register-review passes caught
invented external facts and cross-section contradictions, which is why
that pass stays on specs once per revision (memory
`spec-prose-review-rule`).

The tasks the spec routes into the just-in-time plans, moved here from
`docs/STATUS.md` at the 8c fold: rung 6 (settings, overrides, and
preferences tables, collection-row attributes, device credentials,
quota counters, move journal); rungs 7 and 9 (template grammar and
root-confined writes, audio key by decode plus file hash, break-link
rule, intake caps, same-device check); rung 12 (capability report
extensions and three pinned-build verifications); rung 14 (two mounts,
admin endpoints including seize and removal, re-auth on sensitive
settings); rung 16 (single-link migration, restore-size runbook line);
and the 2.0 projector conformance list. The spec's own "What this spec
routes elsewhere" section is the authoritative list.

## Rung 5, pass 5a (the serve shell and ingress): closed 2026-09-08, lane 2

Executed against
`docs/superpowers/plans/2026-09-08-rung-5a-serve-shell.md` (revision 2)
in the worktree at `.claude/worktrees/lane-2-5a`, branch `lane-2-5a`
from `8ac38d7`. Tasks 1 through 7 ran through the pass-execute chain
(`wf_115f4cec-9c0`, 18 agents, 1.68M); this entry is Task 8, the close.
Task 0 (2026-09-08) found no blockers: a stale lane-registry text, a
kind-constants pointer resolving to `thresholds.go:16-21`, the tracked
incumbent `systemd/dubplate-serve.service` making 5d's unit a
replacement rather than a birth, and a gate time near five minutes
cold. Task 0 also found the spec's 5a row stating a four-rule
normalization where the plan lands six, corrected below.

### What landed

Task 1 (`bbf9b05`) added `internal/serve/clientaddr.go`:
`TrustedProxyConfig`, `PeerAddress` (`RemoteAddr` with the port
stripped, the sole input to every trust decision), `ClientIP` (the peer
address unless a configured header arrives from a configured proxy,
multi-hop values parsed right to left), and the `ProxyMode` descriptor
5d's `doctor` row renders (R27).

Task 2 (`df811a5`, fix `bba4422`) added `internal/serve/strip.go`: the
three-group ingress strip. Group 1 removes every identity-header
spelling in dash and underscore form, both derived from one data
declaration (`identitySpellings`) and, after the fix, cleared from
`r.Header` and `r.Trailer` on the same pass (R-P)—the fix corrected a
gap the strip's own reviewer found, where the announced trailer key's
value cleared only at end-of-stream, leaving the key standing before
the body was read. Group 2 removes the forwarded-address family unless
trusted-proxy mode is on and the peer matches, reading only Task 1's
`PeerAddress` and `TrustedProxyConfig`. Group 3 removes every header a
request's own `Connection` header lists, with the upgrade (R-O) and
keep-set (R-S) carve-outs. The exemption mechanism (R-Q) ships
unreachable at rung 5, since no front exists yet. The extras list is
verified name by name against Authelia's, Navidrome's, oauth2-proxy's,
and Cloudflare Access's own documentation, with citation and read date
recorded beside each name.

Task 3 (`d94fe4b`, `opus`) added `internal/serve/normalize.go`:
`Normalize`'s six-rule refusal, wrapping the mux rather than
registering as a route since `http.ServeMux` cleans and redirects
before any route runs (measured: `/dubplate/../x` and `/dubplate/./x`
both answer 307). Rule 1's predicate is
`r.URL.EscapedPath() != r.URL.Path` (R-D), a blanket ban on
percent-encoding rather than a `RawPath`-based test, because `RawPath`
is populated only when the escaped form differs from `Path`'s default
encoding and is empty for an unencoded request. Rule 3 is a
doubled-separator rule, not an empty-segment rule, admitting a single
trailing separator. Rule 5 answers `414` on a path past 2048 bytes,
inclusive (R-H); every other rule answers `400` (R-T). Rule 6 refuses a
request target whose path does not begin with a separator, admitting
an absolute-form target while refusing a `CONNECT` request's empty
path (R-U). The task carried two notes into Task 7's commit: rule 4 is
defense in depth behind rule 1, and the 2048-byte bound is inclusive.

Task 4 (`aaa8975`) added `internal/serve/prefix.go`: `Prefix`,
validated at construction against a package sentinel; `Owns`, a
segment-aware, case-sensitive-bytes match refusing the salvage tree's
`mountBasePath` gap; and `Route`, composing `Owns` with a 404
`Terminator` so an unregistered path under the prefix never reaches the
fallthrough.

Task 5 (`9388926`, fix `99d60e7`) added `internal/serve/serve.go`: the
root composition. `Shell` composes the strip, normalization, prefix
routing, and the fallthrough into the one `http.Handler` C1 requires,
with `Register` (subtree by name, panics on collision) and
`SetFallthrough` (one slot, panics on a second call), a structured
refusal-log record proving the strip ran before normalization, and the
session-cookie removal at the fallthrough boundary (R-E). The fix
addressed the second review's findings: `statusCapture` gained
`Unwrap`/`Flush`/`Hijack` so `http.ResponseController` and direct type
assertions reach the real `ResponseWriter` through `logRefusal`'s
wrapper; `removeCookie` moved from rebuilding the `Cookie` header via
`http.Request.Cookies()` to editing the raw header text directly, so a
surviving cookie's quoting and any header `net/http` would otherwise
consider invalid pass through verbatim (this closed the multi-valued
`Cookie` finding the second review filed); and a doc-comment
correction to `logRefusal`.

Task 6 (`322a8fc`) added `internal/serve/server.go`: the four `C6`
timeout constants (read-header, read, idle, default write) and
`MaxHeaderBytes`; the `IP_FREEBIND` listener set via
`syscall.SetsockoptInt` inside a `net.ListenConfig.Control` hook (R-M),
the constant verified against the kernel header rather than carried
from memory; and `Run(ctx, ...)`. The default write timeout stays
overridable per request through
`http.ResponseController.SetWriteDeadline`, already exposed to the
fallthrough by Task 5's `statusCapture`. The second review escalated
two findings that the post-chain fix round closed rather than a
same-task fix: `defaultWriteTimeout` as a mutable package variable
(fixed by threading the write timeout through `newHTTPServer` as a
parameter) and `serveUntilDone` not shutting the server down when one
listener's `Serve` failed (fixed by shutting down the whole set on any
one failure).

Task 7 (`799419a`) declared `DUBPLATE_SERVE_BIND_ADDRS`,
`DUBPLATE_SERVE_PORT`, and `DUBPLATE_SERVE_PREFIX` in
`cmd/dubplate-server/serve.go` and deleted `internal/config/keys.go`'s
`serve` reserved class in the same commit, the removal half of the
rung 1b registry contract's first exercise. It wired `newServeCmd`
into `newRootCmd`'s `AddCommand` block, validating the owned wall
through `config.Load` before any listener opens. It moved the
template-classification test to `cmd/dubplate-server`, where the
production registry is composed as the binary builds it, since the
rung 1b registry contract means `internal/config` alone can no longer
classify a template's names once a later rung's keys live only in its
own file; `internal/config` keeps a generic classification-mechanics
test against a fixture registry it declares for itself. It taught the
wall's fifth classification: a registered threshold is a declared
name, so `rejectNamespace` (via `thresholdOf`) accepts a value set
against it rather than reporting it unknown. This task first ran red
and uncommitted: deleting the `serve` reserved class broke
`internal/config`'s own template-classification test, the consequence
the move to `cmd/dubplate-server` then resolved.

The post-chain fix round (`0d3c644`, `opus`) folded the two chain
reviews' escalations together: the trailer scrub arms unconditionally
on the request rather than on the announcement, since a chunked
request may send a trailer section it never announced; the group 3
keep-set compares on header names alone (dropping the session-cookie
name, which is a cookie and not a header) and the configured
trusted-proxy header is exempt from group 1's own removal so a
client-supplied `Connection` cannot delete a trusted front's own
assertion; the session cookie is removed from every `Cookie` header
line rather than the first, by raw-segment splitting; the refusal
record is keyed on the shell's own refusals via a request-scoped
marker, so a subtree answering 400 to its own bad input emits none;
`Terminator` is wired explicitly for the owned prefix;
`io.ReaderFrom` is forwarded through the status capture; every server
timeout became a constant, the write timeout threaded through
`newHTTPServer`; `serveUntilDone` shuts the whole listener set down on
one failure; and the forwarding header is read across every line it
arrived on, right to left as one chain, rather than from the first
line alone.

The fresh-context `web-auth-security-reviewer` (`claude-opus-5`) then
read the whole shell against the plan's and the spec's threat model:
24 findings (1 blocker, 5 majors, 18 minors), do not merge as-is. The
fold (`307001b`, `opus`, cited by the review's disposition record as
"5550e45" before the merge's rebase) closed all six of the
blocker-and-majors plus nine of the eighteen minors: the refusal
record bounds a logged path to a 256-byte prefix and its full length
rather than writing it whole, closing a 1 MiB-per-request flood
vector; `http.Server.ErrorLog` is wired to the shell's own logger at
`warn`; the session-cookie removal's name-matching now mirrors
`net/http`'s own `readCookies` trim order; group 2's removal set is
derived at `NewStrip` as the fixed family plus the configured
trusted-proxy header, and group 3 keeps those names standing for a
trusted peer, closing the "Connection deletes a trusted proxy's own
header" interaction; the trailer scrub gained a `Close` path so a
handler that closes the body without reading it no longer sees a
smuggled trailer; and `main` wires `signal.NotifyContext` plus
`ExecuteContext`, which is what makes the graceful drain built at Task
6 reachable at all outside a test. The nine folded minors also refuse
the wildcard bind address, validate a `Register` subtree name, report
each invalid-prefix shape by its own reason, refuse `TRACE` and
`TRACK` with `405` and no `Allow` header (a stated deviation),
build-constrain `IP_FREEBIND` behind `//go:build linux`, extend group
2 to the CDN client-IP headers (`Cf-Connecting-Ip`, `True-Client-Ip`,
`Fastly-Client-Ip`, `X-Azure-Clientip`), and complete a second citation
pass over the identity-extras list (four names added, one retracted
from the dropped list and cited, four candidates confirmed still
uncitable). Every one of twelve named mutations in the fold's own
review fired. The full findings list, its dispositions, and the
verified-sound table are recorded at
`docs/superpowers/research/2026-09-08-rung-5a-security-review.md`, the
do-not-re-litigate record for 5b, 5c, and 5d.

### The exported surface

`TrustedProxyConfig`, `PeerAddress`, `ClientIP`, `ProxyMode`
(`clientaddr.go`); `Strip`, `NewStrip`, `FrontExemption`,
`NewFrontExemption` and the three groups (`strip.go`); `Normalize` and
its six rules (`normalize.go`); `Prefix`, `NewPrefix`, `Owns`, `Route`,
`Terminator` (`prefix.go`); `Shell`, `NewShell`, `Register`,
`SetFallthrough`, `Handler`, `Logger`, and the refusal record
(`serve.go`); `Listen`, `Run`, the timeout constants, `SplitBindAddrs`,
and freebind (`server.go`, `bindaddrs.go`, `freebind_linux.go`,
`freebind_other.go`); the `serve` verb and its three thresholds
(`cmd/dubplate-server/serve.go`); the moved `templatekeys_test.go`; and
`internal/config`'s fifth wall classification (`thresholdOf` inside
`rejectNamespace`).

### What the gate caught

Task 3's review escalated on notes rather than a fix: rule 4 unreachable
behind rule 1, and the 2048-byte bound's inclusivity, both folded into
the task's own commit. Task 5's second review escalated on a
multi-valued `Cookie` header, `logRefusal` keyed on status rather than
the shell's own refusals, and `Terminator` left unwired for the owned
prefix, all closed in `99d60e7`. Task 6's review escalated on
`defaultWriteTimeout` as a mutable package variable and
`serveUntilDone` not shutting down the listener set on one failure,
both closed in the post-chain fix round rather than in-task. Task 7
ran red once, uncommitted, on the template-classification consequence
described above, resolved in the same task before commit. The
fresh-context security review is the gate's largest catch by finding
count (24) and is detailed above and in the security-review record.

### The resolutions, as ruled

All twenty-two resolutions, R-A through R-V, ruled by the conductor at
dispatch, 2026-09-08, stand as applied. Each row names the file and
symbol carrying the decision.

| Resolution | Ruling | Landed at |
| --- | --- | --- |
| R-A | Owned prefix defaults to `/dubplate` | `cmd/dubplate-server/serve.go:57` |
| R-B | Bind default `127.0.0.1`, port default `8080`, bind list as `KindString` split in-package | `cmd/dubplate-server/serve.go:30`, `internal/serve/bindaddrs.go:38` |
| R-C | Serve keys declared in `cmd/dubplate-server/serve.go`, not `internal/config` | `cmd/dubplate-server/serve.go:13` |
| R-D | Rule 1 refuses on `EscapedPath() != Path`, a blanket percent-encoding ban | `internal/serve/normalize.go:38` |
| R-E | Session-cookie removal lands at 5a, at the fallthrough, cookie name a construction value | `internal/serve/serve.go:219` |
| R-F | Origin check is 5b's; insertion point documented (after strip, after normalization) | Plan only, no 5a code |
| R-G | `doctor`'s client-IP-mode and bound-address rows are 5d's; 5a exports the values | `internal/serve/clientaddr.go` (`ProxyMode`), `internal/serve/server.go:111` (`Addr`) |
| R-H | Explicit path-length bound (2048, inclusive) and explicit `MaxHeaderBytes` | `internal/serve/normalize.go:19`, `internal/serve/server.go:45` |
| R-I | Trusted-proxy configuration is 5d's; 5a's derivation takes plain values | `internal/serve/clientaddr.go:17` |
| R-J | Fresh-context `web-auth-security-reviewer` runs at 5a's own pass end | This entry, security-review record |
| R-K | 5a writes no systemd unit, no deploy path | Ruling only |
| R-L | Extension point registers by subtree name, takes `http.Handler` | `internal/serve/serve.go:80` |
| R-M | `IP_FREEBIND` through `syscall`, not `x/sys/unix` | `internal/serve/freebind_linux.go:10` |
| R-N | Empty fallthrough answers 404 | `internal/serve/serve.go:133` |
| R-O | Fixed hop-by-hop names stay rung 12's; `Connection`-listed removal with the upgrade carve-out | `internal/serve/strip.go:481` |
| R-P | Strip clears identity spellings from `r.Trailer` alongside `r.Header` | `internal/serve/strip.go:268`, `:343` |
| R-Q | Exempted header's duplicate collapses to its last value | `internal/serve/strip.go:350`, `:377` |
| R-R | Group 2 extends to the whole `X-Forwarded-*` family | `internal/serve/strip.go:143` |
| R-S | Group 3's keep-set (`Cookie`, `Origin`, `Sec-Fetch-Site`, session cookie) | `internal/serve/strip.go:187`, `:293` |
| R-T | `400` is the shared refusal status; the length bound alone answers `414` | `internal/serve/normalize.go:80` |
| R-U | Non-origin-form request target refused; absolute-form admitted | `internal/serve/normalize.go:71` |
| R-V | Read timeout adopted at 30 seconds | `internal/serve/server.go:25`, `server_test.go:21` |

### The shell's contract coverage, C1 through C8

| Contract | Go site | Proved by | Reach |
| --- | --- | --- | --- |
| C1 root composition order | `Shell.Handler` (`serve.go:150`) | `TestOrderStripBeforeOwnedSubtree`, `TestOrderStripBeforeFallthrough`, `TestOrderRawVsDecodedRefusedBeforeAnyHandler`, `TestOrderPrefixRequiresSegmentBoundary` (`serve_test.go:69,94,123,160`) | Behavioral, not structural, per the plan's own rule |
| C2 three-group ingress strip | `Strip.Wrap` (`strip.go:333`) | `TestStripWrapRemovesEveryIdentitySpelling`, `TestStripWrapConnectionListedRemoval`, `TestStripWrapGroup2ForwardedAddress`, plus 11 more rows in `strip_test.go` | Full, both header and trailer paths |
| C3 normalization rules | `Normalize` (`normalize.go:83`) | `TestNormalize` (table over all six rules), `TestUnencodedPathCarriesEmptyRawPath`, `TestServeMuxCleansBeforeARouteRuns` | Full for all six rules |
| C4 prefix routing and the terminator | `Prefix.Route`, `Terminator` (`prefix.go:149`, `:131`) | `TestRoute`, `TestNewPrefix`, `TestUnclaimedOwnedPathReachesTheTerminator` | Full |
| C5 the extension point | `Shell.Register`, `Shell.SetFallthrough` (`serve.go:80`, `:136`) | `TestRegisterDuplicateSubtreePanics`, `TestSetFallthroughTwicePanics`, `TestTwoSubtreesEachReceiveOnlyOwnPaths`, `TestRegisterAfterHandlerPanics` | Full for the collision and slot rules; the static two-handler split is unreached, since 5a builds neither handler (declared gap 6) |
| C6 the deadline policy | `newHTTPServer`, the timeout constants (`server.go:18`, `:92`) | `TestNewHTTPServerTimeouts`, `TestWriteDeadlineExtension` | Full for the mechanism; slowloris and slow-body resistance rest on the constants rather than a timing test (declared gap 5) |
| C7 two addresses, two jobs | `PeerAddress`, `ClientIP` (`clientaddr.go:68`, `:102`) | `TestPeerAddress`, `TestClientIP`, `TestClientIPReadsEveryHeaderLineAsOneChain` | Full |
| C8 the bind surface | `Listen`, `SplitBindAddrs`, freebind (`server.go:113`, `bindaddrs.go:38`, `freebind_linux.go`) | `TestListenBindsExactAddresses`, `TestControlHookOpensAnUnconfiguredAddress`, `TestSplitBindAddrs` | Full for binding; the `doctor` row surfacing a misconfigured address is 5d's (R-G) |

### The mutation table

Named mutations as the plan predicted them per task; an outcome
attributed beyond the plan's own text is reported by the implementer
or the review record cited above.

**Task 1** (5 named): reading the header without checking the peer, a
left-to-right multi-hop parse, an empty-header fallback, moving the
header check into the peer-address function, and hand-splitting
`RemoteAddr` on the last colon all fire on their named rows. Mutation
4 is recorded as the one whose blast radius crosses into Tasks 2 and 5.

**Task 2** (10 named): dropping the underscore derivation, deleting
after the handler returns, skipping group 2 when mode is off, dropping
the `Connection`-listed removal, exempting the configured header for
every peer, reading the peer address from `X-Forwarded-For` inside the
strip, dropping the trailer clause, removing every `Connection`-listed
name unconditionally, taking the exempted header's first value instead
of its last, and dropping the `X-Forwarded-*` family all fire on their
named rows. Mutation 6 is the one Task 1's mutation 4 predicted, fired
here.

**Task 3** (10 named, `opus`): comparing `RawPath` against `Path`
instead of `EscapedPath()` fires the plain-path row, the R-D trap
proven real. Cleaning `..` instead of refusing, dropping the backslash
rule, dropping the doubled-separator rule, registering the check as a
route instead of wrapping it, varying the refusal status per rule,
refusing an empty segment instead of a doubled separator, dropping the
origin-form rule, admitting a benign encoded byte, and dropping the
`.` segment from rule 2 all fire on their named rows.

**Task 4** (5 named): `strings.HasPrefix` in place of the segment-aware
match, lowercasing before matching, falling through instead of
terminating, terminating with the wrong status, and accepting a prefix
with a trailing separator at construction all fire on their named
rows.

**Task 5** (8 named): swapping the strip and normalization steps,
registering normalization as a route, reordering the prefix match
ahead of normalization, letting the fallthrough see a refused request,
allowing a silent duplicate registration, removing the session cookie
on every path instead of at the fallthrough, deleting the whole
`Cookie` header at the fallthrough, and stripping after the prefix
decision instead of before all fire on their named rows.

**Task 6** (7 named): dropping the write deadline, dropping the
`ResponseController` extension, binding `0.0.0.0` instead of the
address list, returning on the first bind error without closing the
openers, and dropping the `Control` hook all fire on their named rows;
the `Control`-hook mutation is the one revision 1 pre-authorized as
unfired on the belief every test address was already present, which
this task's own execution corrected. Taking `ctx` off `Run` fails to
compile, the mutation the language itself catches. Dropping the read
timeout is **reported unfired by design**: a slow-body test is a
timing test this gate does not carry, and the constant's presence in
the server literal, with its reason comment, is what holds the
property instead (declared gap 5).

**Task 7** (6 named): declaring the keys without deleting the reserved
class fires the registry collision panic; deleting the class without
declaring the keys fails the declaration test; giving
`DUBPLATE_SERVE_PREFIX` an empty default panics at
`validateThresholdDefault`; wiring `newServeCmd` outside the
`AddCommand` block fails the verb's invocation test; passing
`*config.Config` into `internal/serve` is caught by the reviewer's read
of the import list. Registering `DUBPLATE_SERVE_PORT` as `KindString`
is **reported unfired by design**: the default `"8080"` still parses as
a string and no test fails on the kind alone, held instead by the
declaration's stated kind, read at review.

**The security fold's own review** named twelve mutations against its
six fixes; every one fired, reported by the fold's own review record.

### The declared coverage gaps

**Six declared by the plan's own Task 8 template, joined by five from
the security review's residual findings, eleven new gaps against the
running total of fifty-two after rung 8b (rung 3's twenty-one, rung
4a's nine, rung 4b's four, rung 1b's two, rung 8a's seven, rung 8b's
nine), bringing the total to sixty-three.**

1. **The strip's extras list is complete only to the documentation each
   name was verified against.** A spelling no reference implementation
   documents is a spelling the set does not carry; the mitigation is
   that a new one is one line of data.
2. **The exemption mechanism is unreachable at rung 5** and is tested
   against a synthetic front. No real forward-auth front exists to
   grade it until one is built.
3. **No `doctor` row reports the client-IP mode or the bound
   addresses** (R-G). 5a exports both; 5d renders them.
4. **The session-cookie removal is tested with an arbitrary cookie
   name** (R-E), because the real name arrives at 5b.
5. **Slowloris resistance and slow-body resistance rest on the timeout
   constants rather than on a test**, since a byte-at-a-time client is
   a timing test this gate should not carry.
6. **No page, no auth handler, and no real consumer has registered at
   the extension point.** 5b is the first real registration.
7. **`identity_spelling_present` proves order, not value** (security
   review finding 8). Go populates an announced trailer key with a nil
   value before the handler reads anything, so the refusal-record
   assertion reads key presence rather than content; sound for C1's
   ordering claim, not a value-content proof.
8. **No connection cap alongside the 120-second idle timeout**
   (security review finding 14). The idle timeout is the stated
   mitigation for how long one connection may sit; nothing bounds how
   many a client holds open at once.
9. **A `;`-separated path segment reaches the fallthrough unrefused**
   (security review finding 17). Neither C3's rules nor
   `http.ServeMux`'s own cleaning refuse it, so a router built to
   interpret it differently than dubplate's own routing can disagree.
10. **A case-varying path is routed, not refused, at the prefix
    boundary** (security review finding 18). C4's case-sensitive match
    is the stated policy; the residual is that the mismatch is silent
    traffic on the fallthrough rather than a refusal.
11. **A registered threshold is accepted by `Load` with any value**
    (security review finding 24). The wall's fifth classification
    accepts a threshold-classified name regardless of the value
    supplied; no per-threshold value validation runs beyond the
    registry's own default-presence check, a property of the rung 1b
    wall contract rather than a 5a-created gap in the code this pass
    wrote.

Six review findings are recorded as forward-scope carries rather than
gaps this pass created, since the surface each names does not exist
yet: finding 7 (absolute-form `r.Host`, 5b), finding 16 (no
`Referrer-Policy`, 5c), finding 22 (`netip`-typed proxy set, 5d), and
finding 23 (`check`/`doctor` value validation, 5d). These are listed in
the handoff and the security-review record, not counted twice here.

### The 5a to 5b handoff

5b inherits: the composed root handler it mounts auth handlers into
through `Shell.Register`; the extension point's registration method
and its collision panic; Task 1's `ClientIP` derivation, which the
three limiter keys read and which 5b must not re-derive; the
constraint that the origin check inserts after the strip and after
normalization (R-F), so a smuggled path never reaches it and no
inbound identity header is present when it decides; the session-cookie
name the shell already takes at construction (R-E); the keep-set that
guarantees `Origin` and `Sec-Fetch-Site` survive ingress (R-S),
alongside the fact that Go's `CrossOriginProtection.Check` returns nil
when both are absent; the 30-second read timeout and the
`http.MaxBytesReader` alternative if a later rung wants a streaming
upload (R-V); R33's disposal as 5b's post-login destination rule
(`spec:674-686`, not Task 3's); rule 1's blanket percent-encoding ban
as a named strictness rather than a permanent default, revisited at
rung 12 against the streamer's real client paths with measured
evidence before relaxing it; and the fact that `internal/config`'s
`keys.go` still carries the `auth and identity` class and the pending
wall row, both untouched and both 5b's to retire.

The security review adds three requirements to that inheritance: 5b
builds every URL from the configured public origin, never `r.Host`
(finding 7); `Referrer-Policy: no-referrer` and the rest of the served
header set are 5c's, not 5a's (finding 16); and the moved
`templatekeys_test.go` classifies against the production registry as
composed in `cmd/dubplate-server`, which 5b's own new keys join rather
than re-derive.

### Forward carries

1. **The `netip`-typed trusted-proxy address set** (security review
   finding 22) is 5d's, when the trusted-proxy integration class gives
   the configured set real shape.
2. **`check` and `doctor` gain no serve-value validation at 5a**
   (security review finding 23); 5d's `doctor` row (R-G) is where a
   configured bind address or prefix gains a live-evidence check.
3. **`Referrer-Policy: no-referrer` and the rest of the served header
   set** are 5c's (security review finding 16).
4. **The `;`-segment and case-variant-prefix fallthrough gaps** (coverage
   gaps 9 and 10 above) are rung 12's to close when a real driver gives
   them a router to disagree with.
5. **`internal/config`'s `auth and identity` reserved class, the
   pending identity-source wall row, and the three supersession-message
   rewrites** are untouched by 5a and are 5b's whole edit to
   `keys.go` and `config.go`, per the contested-surfaces table.

### The spec corrections

Applied directly to
`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md` as dated
errata under "Errata the close applies":

1. **R23's ceiling table**: 5a's ceiling re-derived from rung 8a's
   actual, 3M to 4.25M; rung 5's combined reservation moves from
   14.5M to 15.75M.
2. **R24's review shape**: gains 5a's own fresh-context
   `web-auth-security-reviewer` read (R-J), argued from the spec's own
   split-point paragraph.
3. **The session-cookie removal's pass assignment** (R-E): settled as
   5a's, at the fallthrough, parameterized by the cookie name.
4. **The hop-by-hop scope** (R-O): narrowed to the `Connection`-listed
   removal with the upgrade and keep-set carve-outs; the fixed RFC 7230
   set stays unremoved by 5a's own code, on `httputil.ReverseProxy`
   already removing it.
5. **The strip's trailer coverage** (R-P): the claim list now names
   both `r.Header` and `r.Trailer`.
6. **Group 2's `X-Forwarded-*` extension** (R-R): the fixed three-name
   list grows to the whole family.
7. **Rule 1's mechanism and rule 3's shape** (R-D, and the composition
   order's "any empty segment" corrected to a doubled-separator rule
   with a single trailing separator admitted). **The normalization
   rule count moves from four to six**, correcting `spec:2058`'s pass-
   split row.
8. **The read timeout** (R-V): a fourth timeout named alongside the
   three the deadlines section already carried.
9. **The template-classification test's home and the wall's fifth
   classification**: the test lives in `cmd/dubplate-server`, a
   consequence of the rung 1b registry contract; `rejectNamespace`
   gains the registered-threshold branch.

### The evidence

`go test ./internal/serve/... -count=1` is green in 1.130s-1.143s
across repeated runs. `go test ./... -count=1` (every package) is
green in 2m33.785s, no existing suite's result changed by this pass.
The strip's extras list, final: `Remote-User`, `Remote-Name`,
`Remote-Email`, `Remote-Groups`, `X-Forwarded-User`,
`X-Forwarded-Email`, `X-Forwarded-Preferred-Username`,
`X-Forwarded-Groups`, `X-Forwarded-Access-Token`, `X-Forwarded-Uri`,
`X-Forwarded-Method`, `Cf-Access-Jwt-Assertion`,
`Cf-Access-Authenticated-User-Email`, `X-Webauth-User`, the five
`X-Authentik-*` names, `Tailscale-User-Login`, `Tailscale-User-Name`,
and `X-Original-Url`, each dash-and-underscore, cited against Authelia,
Navidrome, oauth2-proxy, Cloudflare Access, Grafana, Gitea, authentik,
and Tailscale's own documentation; `X-Client-Ip`, `X-Cluster-Client-Ip`,
`X-Original-Method`, `X-Rewrite-Url`, Miniflux's `AUTH_PROXY_HEADER`,
oauth2-proxy's `X-Auth-Request-*` set, and Cloudflare's service-token
headers stay dropped, each reason recorded beside it in
`internal/serve/strip.go`. `internal/config` and `cmd/dubplate-server`
each carry only the touch this pass's Global Constraints named: the
`serve` reserved-class deletion and the templatekeys-test move,
nothing else in `internal/config`.

### What a later pass would be wrong to rediscover

`RawPath` is never the basis for the percent-encoding rule: it is
populated only when the escaped form differs from `Path`'s default
encoding, so a `RawPath`-keyed test passes over an ordinary path by
accident. The trailer scrub must run on both the read path and the
`Close` path; a handler that closes the body without reading it is a
real request shape, not a hypothetical one. A client's `Connection`
header can delete a trusted front's own forwarded-address assertion
unless group 3's removal explicitly excludes group 2's kept names for
a trusted peer; the two groups are not independent. `net/http`'s own
`readCookies` trims a cookie name after the `=` cut, not before; a
cookie-removal routine that trims first admits a name carrying a
trailing space as a distinct cookie from the one it means to remove.
The graceful drain `Run` builds is inert without a signal-driven
context in `main`; a shutdown mechanism with no caller that cancels
its context is not shipped, whatever its own tests prove. The
template-classification test's home is a permanent consequence of the
rung 1b registry contract, not a 5a-specific choice: any later rung
declaring its own keys in its own file, as rung 5 does, moves the
"every template name classifies" proof to `cmd/dubplate-server` with
it, since `internal/config` alone can no longer see the whole
registry.

### Deviations

The conditional `opus` upshift named for Task 5 (if Task 3's review
returned `fix`) was not applied: Task 3's review escalated on notes,
not a `fix` verdict. The post-chain fix round and the security fold
both ran on `opus` by conductor choice rather than the plan's own
per-task model assignment. The refusal marker is a request-context
value rather than a capture-field, a shape difference from how earlier
rungs held similar per-request state. `Run` gained a logger parameter
beyond what the plan's own signature draft named.

### Spend

About 3.9M against the 4.25M ceiling: plan preparation, 0.83M; Task 0,
0.13M; the chain, 1.68M across 18 agents; the post-chain fix round,
0.27M; the security review, 0.18M; the security fold, 0.22M; the merge
review, 0.09M; arguments and commits, 0.3M; conductor turns, about
0.2M. The 3.4M flag was crossed by the security fold, after every task
had already landed. Attended interactions: one planning miss (the
template-classification consequence of rung 1b's contract, which a
planning-time probe of `config_test.go` would have caught), zero
execution sittings.

### The merge ritual

Rebase onto master clean (master had moved by 8b's own close); full
gate green in the worktree; `--ff-only` merge to `307001b`; full gate
green on master; one `diff-reviewer` over the merged range, since the
rebase was not a fast-forward: accept, with three carries for 5b
(`templatekeys_test.go` classifies a pending wall row as owned where
`Load` would refuse it; `productionRegistry` restates `newRootCmd`'s
register calls by hand; `server.go:161` passes the listener as a
goroutine parameter, a pre-1.22 idiom) and the note that the arch
tests root at `internal/engine`, so `internal/serve` sits outside them
by design. The second code slot released; the worktree torn down.

## Rung 8, pass 8b (PICTURE and the conformance instrument): closed 2026-09-08

Tasks 0 through 6 executed against
`docs/superpowers/plans/2026-09-08-rung-8b-picture-conformance.md`
(revision 2) in the main tree. This entry is Task 7, the pass close.
Task 0 found no blockers: 8a's simplifier round at `0f2b61d` moved no
citation, `.gitleaksignore` stayed tracked and empty, and
`adjudication.go`'s required-field checks sit at `:201-215` as the plan
assumed. Task 0 also flagged `docs/HISTORY.md:442-447`'s forward-carry-5
pointer as stale (corrected below, at the rung 4b entry's own item 5)
and noted gitleaks now also scans lane 2's worktree from the main
tree's run, about three minutes added to that leg.

### What landed

Task 1 (`a4c2bcf`) added `Picture` (`internal/engine/tagio/picture.go:40`)
and widened `Container` to decode every PICTURE block in the chain
(`Container.Pictures`, `container.go:112`): all eight RFC 9639 section
8.8 values, bounds-checked against the block's own payload extent
before a slice is taken, an out-of-range type index degraded to 0, and
`FrontCover()` (`picture.go:76`) reproducing the single-candidate-then-
front selection. The fix in the same commit repaired an 8a
write-round-trip fixture that used arbitrary bytes for block type 6,
which real PICTURE decode now rejects.

Task 2 (`1b8bd85`) widened `Write` (`write.go:39`) with the ordered
picture set between the tag set and the file mode: R1's three caller
states (nil leaves pictures untouched, a non-nil empty slice strips,
a non-empty slice replaces), P7's sniff-overrides-declared media type
over R2's six-row recognized set, and P6's zeroed geometry on every
write. PICTURE joined VORBIS_COMMENT as a block type the writer owns:
`planBlocks` (`write.go:121`) stopped carrying incoming PICTURE blocks
through, a rewritten picture replaces its own block in place, and a
picture written into a container with none lands after VORBIS_COMMENT
and before the trailing PADDING. `deviations.go` gained four rows
(P6's zero geometry, R3's undetected APNG, R2's CR2 refusal, and P3's
block-length trust, the last covering VORBIS_COMMENT as well and
recording a policy 8a already applied silently). The widened signature
touched sixteen call sites, all inside the package's own tests.

Task 3 (`a803ba1`, fix `ae37b5e`) added the seam registry
(`seam.go:50`, keyed by the exported `Format`, `seam.go:21`) so `Read`
and `Write` sniff, look up and dispatch through one unexported
`map[Format]formatEntry` of func values rather than an interface
(ruling 14). `Parse` deliberately bypasses the registry (ruled, m6):
it stays the FLAC-specific entry point, and the emptied-registry test
proves `Read` and `Write` stop reaching the golden while `Parse` keeps
parsing it. The fix documented the bypass on `Parse`'s own doc comment
and in the package doc.

Task 4 (`f11e6fa`, fix `b0d2107`) added R11's JSON tags to all 69
`TagSet` members (`tagset.go`), each naming the ledger's own field
name with no `omitempty`, tied to `Field` two ways (by name through
`fieldStructName`, by position against `Mapping()`), and landed
`TestConformanceLedgerSubjects` (`conform_test.go:640`) through P10's
first three steps: the twelve subjects derive from
`corpus/analysis/file_tags.json`'s own rows (R6), each subject's
applied fields are filtered to the ledger's own `writable_fields`
header, and the filtered object fills a `TagSet` under the four-state
rule and writes through `Write` into a per-subject scratch copy of the
template (R5's path constants, local decode types per the arch walk).
The fix repointed a doc comment from the wrong test name to
`TestTagSetJSONTagsTieToFieldConstants`, the one that actually ties the
tags.

Task 5 (`d824be5`, `opus`) landed P10's steps 4 and 5 and the exit
gate's proof that the instrument can fail. Each subject's scratch file
is read back through `tagio.Read`, the package's one read entry point,
and projected to the ledger's shape through R11's tags, so no second
projection exists. `compareSubject` (`conform_test.go:257`) grades the
union of both sides' keys under R10's exact typed comparison: both
sides decoded through `encoding/json`, equality exact with no
tolerance on any float, a list compared in its own order, and a key
present on one side alone a divergence rather than a silent non-
comparison. `TestCompareSubjectNamesEveryDisagreement` discharges the
exit gate's fail-proof requirement as a standing table over synthetic
rows rather than a one-off demonstration. **All twelve subjects agree
on all 69 keys**, so nothing was folded into
`corpus/analysis/adjudications.json`.

Task 6 (`f146623`) closed `RE:192` (rung 3 had a tier-2 pending-row
escape, rung 8 had none): `AdjudicationEntry` gained an optional
`AppliedField` block (`internal/harness/analysis/adjudication.go:92`,
type at `:114`) carrying the item index, the field name, the ledger's
value, the writer's read-back value, and the `tagio` commit, legal only
on a case-subject row. `internal/engine/tagio` builds the same shape
over a local struct (`conform_test.go`, disposition and adjudicator
spelled as literals) because it may never import the harness; the two
packages are tied only through the committed example both grade
byte for byte and by decode against
(`internal/engine/tagio/testdata/divergence-example.json`).

The leftovers round (`a70b346`) closed four findings without a second
reviewer, under the window-cap ruling: `SupportedFormats` now returns
registry keys sorted by `Format` value rather than a map's own
iteration order, and `read.go`/`write.go` nil-guard the registry row
lookup, answering `ErrUnsupportedContainer` instead of panicking on an
absent or func-less row; `adjudication.go`'s "no matcher stamp on an
applied-field row" rule became its own check rather than leaning on the
general case-subject check, proven by a mutation that removes it and
watches the new check's own message disappear while the older, general
check still fires; the golden PICTURE block's sum-to-declared-length
assertion now compares against the literal 78 rather than
`c.Blocks[2].Length`, so it cannot pass by agreeing with itself, a
refusal row was added for a payload carrying no geometry, and
`Picture.Data`'s aliasing of the decoded block's own payload is now
documented on the field; and `conform_test.go` dropped a restated
divergence-route paragraph in favor of citing
`analysis.AdjudicationEntry.AppliedField`, marked `instrumentDate` as
the committed example's fixed date rather than a clock read, simplified
`*ts.Comp != false` to `*ts.Comp`, and swapped the committed template's
before/after guard from a size comparison to a SHA-256 content hash.

The batched `code-simplifier` round (`ce0ac2e`) ran once over 8b's Go
after Task 6, per the plan's single named round: it extracted
`resolveEntry` (`seam.go:112`) so `Read` and `Write` share the same
sniff-then-lookup dispatch instead of repeating it.

**The exported surface added this pass**: `Picture`
(`picture.go:40`), `Container.Pictures` and `FrontCover()`
(`container.go:112`, `picture.go:76`), `Write`'s picture argument and
its three states (nil, non-nil empty, non-empty), the sniff and its
six recognized formats (`picture.go:172`), `Format` and
`SupportedFormats()` and the unexported registry (`seam.go:21`, `:71`,
`:50`), the conformance instrument (`conform_test.go`) and its
comparator `compareSubject`, and
`analysis.AdjudicationEntry.AppliedField` with the divergence example
it and the instrument's emitter both grade against.

### What the gate caught

No task needed a second `diff-reviewer` round beyond Task 3's single
fix and Task 4's single fix, both targeted documentation corrections
rather than behavioral defects. The leftovers round's four findings
(above) are review nits and one hardening (the nil-guard on an absent
registry row), not defects a mutation caught unaware; each carries its
own proof in the commit body rather than a fresh adversarial pass, per
the same window-cap reasoning 8a's own leftovers round used.

### The resolutions and rulings, as applied

All twelve Resolutions stand as ruled, conductor 2026-09-08, per the
plan's Resolutions section; none folded.

- **R1** (three caller states: nil leaves pictures untouched, non-nil
  empty strips, non-empty replaces): `write.go:39`'s argument list,
  proven by Task 2's three-state tests.
- **R2** (the six-row recognized set, reproducing `filetype`'s whole
  test rather than its leading prefix, and a CR2-prefixed blob
  refused): `picture.go:172` (`sniffMediaType`), the CR2 row in
  `deviations.go`. Ruled otherwise, a seventh sniff row would reproduce
  `image/x-canon-cr2` at the cost of a format no cover realistically
  carries.
- **R3** (APNG discrimination not reproduced, PNG's signature alone):
  `picture.go`'s sniff, the deviation row in `deviations.go`. Ruled
  otherwise, Task 2 would add the chunk walk and its own bounds checks.
- **R4** (picture fixtures built in-test, no new binary committed): no
  file joined `testdata` this pass; every fixture in `picture_test.go`
  and `write_test.go` is a hand-written literal.
- **R5** (the instrument's three tracked inputs named as relative path
  constants, no conditional skip): `conform_test.go`'s path-constant
  block.
- **R6** (the twelve subjects derive from the ledger's own rows, not a
  second hard-coded list): `TestConformanceLedgerSubjects`
  (`conform_test.go:640`).
- **R7** (the schema addition on `AdjudicationEntry`, the instrument's
  own emitter over a local struct taking date and commit as parameters,
  the conductor's hand-fold as the route, and the committed example as
  the anchor against drift): `adjudication.go:92`, `:114`;
  `testdata/divergence-example.json`.
- **R8** (the `io.ReaderAt` gap declined at 8b): no task widened
  `Parse`'s signature; the gap stays declared, below, now named as
  8c's call alone.
- **R9** (the registry row carries the magic prefix beside its func
  values, a struct rather than an interface): `seam.go:33`
  (`formatEntry`), `:50` (`registry`).
- **R10** (typed comparison, exact equality, no tolerance introduced):
  `compareSubject` (`conform_test.go:257`). Task 5's mutation report
  (below) is this resolution's sharpest proof: a tolerance mutation
  fired, correcting the plan's own "no test fails" prediction, so
  R10's rule catches more than the plan expected.
- **R11** (JSON tags naming the ledger's own field names, no
  `omitempty`, tied two ways): `tagset.go`'s 69 member tags,
  `tagset_test.go`'s `TestTagSetJSONTagsTieToFieldConstants` and
  `TestTagSetJSONTagsCarryNoOmitempty`.
- **R12** (`Read` stays tag-only; the picture read reaches through
  `Parse`'s `Container.Pictures`; the double parse is a stated
  deviation): `read.go:17`, `picture.go:76`. Ruled otherwise, Task 1
  would have added a sibling read entry point returning both the tag
  set and the pictures from one parse, and `Read` would become a thin
  wrapper over it; neither this pass's exit gate nor rung 9's arrival
  forced that shape, so the call site is left to pick it, per the
  spec's own correction (below).

### The mutation table

Every task's named mutations, as the plan predicted them and as the
implementer reported running them; outcomes beyond a commit message
are attributed to the implementer's own report.

**Task 1** (7 named): the endianness, bounds-check, type-index-degrade,
and description-length mutations fire on the golden's decode and the
refusal table; decoding only the first PICTURE block fails the
two-block fixture. Mutation 4 (returning the first picture rather than
the first front cover) fires the back-cover-then-front-cover row.
**Mutation 5** (applying the front-cover rule before the single-
candidate rule) has no reproducible form against the single-candidate
shape as literally described: the implementer constructed a faithful
variant to exercise the ordering, which fires the lone-back-cover row,
recorded as a coverage-list item below rather than a plain "fired".

**Task 2** (11 named): reported by the implementer as run per the
plan's predictions, covering the declared-versus-sniffed media type,
non-zero geometry, the three caller-state mutations, the append-versus-
replace mutation, PICTURE's ownership-transition mutation (leaving it
in `planBlocks`'s preserve-unknown path), the little-endian mutation
against the golden's big-endian bytes, the unsniffable-blob mutation,
the two prefix-only-sniff mutations (RIFF without `VP`, TIFF without
the CR2 exclusion), and the chain-order mutations. Mutation 11
(dropping one of `deviations.go`'s four rows) is reported unfired by
design, the same reviewer's-read shape 8a's Task 8 used.

**Task 3** (no named-mutation list of its own; proven by the
emptied-registry test): emptying the registry stops `Read` and `Write`
from reaching the golden while `Parse` keeps parsing it, which is the
proof the registry bypass is a decision rather than an oversight (m6).

**Task 4** (8 named): the tag-mutation, `omitempty`-addition, and the
three four-state mutations (null-as-empty-string, `[]`-as-absent,
zero-as-absent) all fire on the tie assertion or the four-state table.
The writable-filter-skip mutation fires the count assertion at 99
rather than 69. Mutation 7, sharing one scratch path across the twelve
subjects, fires the distinct-path assertion here (the path-level half);
Task 5 carries the same mutation's value-level half. Mutation 8
(hard-coding the twelve subject ids) is reported unfired by design,
held by the reviewer's read of R6's diff.

**Task 5** (7 named): the own-fixture, key-coverage, untouched-template,
null-versus-empty, and list-order mutations all fire as predicted.
**Mutation 6**, the value-level half of the shared-scratch-path leak
(subject N's values appearing in subject N+1's read-back), is
**unfirable by construction**: `Write` rewrites or deletes every owned
key on every call, so a shared path never actually carries a stale
value forward for the comparator to catch, whatever the path-level half
(Task 4) already proves about the paths themselves. **Mutation 7**, introducing a float tolerance, is reported firing at a
tolerance of about 1e-6, correcting the plan's own prediction of "no
test fails": the corpus carries real precision differences a coarse
tolerance would mask, so R10's no-tolerance rule catches more than the
plan expected, never less.

**Task 6** (7 named): the disposition-mutation, shadow-subject
mutation, missing-commit mutation, key-order/indentation mutation, and
no-divergence-no-row mutation all fire on the byte-equality assertion
against the committed example or on the new checker rules. The
relax-the-new-rule mutation fires the malformed-row table's first row.
Mutation 7, stamping the date from the clock inside the instrument,
fires `TestEngineNeverReadsTheClock`, the constraint R7's date-and-
commit parameterisation exists for; the task reports it fired by
design.

**The leftovers round's own mutation** (adjudication.go's stamp rule):
removing the new "no matcher stamp on an applied-field row" check makes
the new subtest fail on a different message (the older, general
case-subject check's), which is the proof the new check is its own
rule rather than borrowed; restored, it passes again.

### The declared coverage gaps

Nine gaps this pass creates, joined to the running total of
forty-three (rung 3's twenty-one, rung 4a's nine, rung 4b's four, rung
1b's two, rung 8a's seven; `docs/HISTORY.md`'s rung 8a entry), bringing
the total to **fifty-two**. Two items this pass's own list carries are
not counted again here because they restate an already-counted 8a gap
rather than declare a new one: "no real library file has been graded"
restates 8a's gap 6, and "the `io.ReaderAt` gap" restates 8a's gap 7,
now narrowed to name 8c alone as the decision-maker.

1. **APNG is reported as PNG** (R3). A stated deviation, held by
   `deviations.go` and by this entry.
2. **The picture set has no oracle-produced multi-picture fixture.**
   The golden carries one picture, so the two-picture and
   front-cover-selection rows are all hand-built.
3. **The sniff covers six formats and the reference covers
   seventeen** (`filetype/types/__init__.py:15-32`). A cover in a
   format outside the six is refused where the reference would have
   written it.
4. **`Read` returns no pictures** (R12), so a caller wanting the tags
   and the cover parses the file twice. A stated deviation from
   `spec:729-733`, handed to rung 9 as the first caller, and now
   corrected at the spec itself (below).
5. **A CR2-prefixed blob is refused** where the reference writes it as
   `image/x-canon-cr2` (R2). A stated deviation, held by
   `deviations.go`.
6. **A block whose declared length disagrees with its contents is
   refused** where mutagen parses it anyway (P3). A stated deviation
   covering VORBIS_COMMENT as well as PICTURE, and the one gap this
   pass records against 8a's landed code as well as its own.
7. **The divergence schema's field name is not checked against the 69
   writable keys**, because the checker does not load the ledger. A
   row naming a field that is not writable passes.
8. **Task 1's mutation 5 (front-cover rule before single-candidate
   rule) has no reproducible form** against the actual single-candidate
   shape; the implementer built a faithful variant rather than the
   literal mutant the plan named, so the ordering proof rests one step
   removed from the plan's own description.
9. **Task 5's mutation 6, the value-level half of the shared-
   scratch-path leak, is unfirable by construction**: `Write`'s
   contract rewrites or deletes every owned key on every call, so no
   scenario exists in which a stale value from a prior subject's write
   survives into the next subject's read-back for the comparator to
   catch.

### The 8b to 8c handoff

8c inherits the widened `Container` and `Write` signatures its absorbed
reader must satisfy: `Container.Pictures` and `Write`'s picture
argument are landed, and the shadow's reader moves against them rather
than against 8a's narrower shapes. 8c's reader routes through the seam
registry the way `Read` and `Write` do; `Parse` stays the one
deliberate bypass (m6), so 8c's absorbed reader is a registry
consumer, not a second bypass. The conformance instrument,
`TestConformanceLedgerSubjects` beside the `-run Conformance` command
rung 15 re-runs (`go test ./internal/engine/tagio/... -run Conformance
-count=1`), must stay green as the shadow's reader moves; a regression
there is a defect in the absorption, not a reason to relax the
instrument. The `io.ReaderAt` call is now 8c's alone (R8): whether
`Parse` grows a second, injectable entry point is 8c's decision, with
no consumer forcing it before then. 8c's plan inherits R12's read-side
answer, and the double parse it accepts, as settled rather than as an
open question: 8c's absorbed reader may keep reading through `Parse`
and `Container.Pictures` directly, since it already holds a parsed
`Container`, and it owes no fold into a combined `Read`. The
tightened refusal set is unchanged by this pass, so 8c's capture-path
work (the tightened refusal ruling 9 names) is written against 8a's
rules, not against anything 8b added.

### Forward carries

1. **Fork B stays open and owed to Geoff**, unchanged by this pass: a
   FLAC-only 1.0 with the MP3 pass immediately after cutover, applied
   provisionally by the conductor per the plan's header.
2. **The group non-FLAC question** (`docs/HISTORY.md`'s rung 8a
   entry) stays carried and is not decided here: this pass builds one
   container's picture layer and the seam that would hold a second,
   and neither answers whether a member's non-FLAC library gets
   a rung 8-shaped pass.
3. **The gitleaks worktree-scan cost** is a `check.sh` follow-on: Task
   0 found gitleaks from the main tree now also scans lane 2's
   worktree, adding about three minutes to that leg. No task in this
   pass owns `scripts/check.sh`; the follow-on is recorded here for
   whichever pass next touches the gate script.
4. **The stale `docs/HISTORY.md:442-447` forward-carry-5 pointer** is
   corrected in this entry's own edit to the rung 4b entry (that
   entry's item 5 now names the mapping's field count correctly and
   points at the 4a plan and `flac.go` rather than the 4a-i entry).

### The spec corrections

Drafted by Task 7 and applied directly to
`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md` as dated errata,
in the same shape the 8a-close erratum used:

1. **The `spec:1103` pointer is stale.** The 8a-close erratum's own
   item (1) cited `spec:1103` for the clause "the seam registry with
   its supported set, its func-value dispatch, and its unsupported-
   container sentinel"; both that erratum's own insertion and this
   pass's own erratum insertion moved the clause, which now sits at
   `spec:1167`, the pass-split table's 8b row. The registry, the
   dispatch and the supported set landed at 8b as that row said; only
   the citation was wrong.
2. **R12's spec correction.** The "two entry points" paragraph's "one
   reads a file's tag set and its pictures" (originally `spec:729-733`,
   now `spec:752-756` after this pass's own erratum insertions moved
   it) is corrected: `Read` stays tag-only, and the picture read is
   reached through `Parse`'s `Container.Pictures`. A caller wanting
   both parses twice, handed to rung 9 as a stated deviation and
   declared coverage gap 4 above, not a defect.
3. **The ceiling table** (originally `spec:1181-1196`, now the
   paragraph at `spec:1214`) is corrected: the 8b row's proposed 3M is
   3.25M, re-derived after Task 2 against this plan's own review. Rung
   8's ceilings now stand at 8.75M against the spec's 8M reservation,
   with 8a's 0.9M underspend as the offset (8a's own close entry,
   above, records 3.6M actual against the 4.5M ceiling it operated
   under, raised from the spec's proposed 4M by the same re-derivation
   rule this correction applies); 8c's proposed 1M line stands
   unchanged, since 4.5M plus
   3.25M plus 1M is 8.75M and 8c is the mechanical absorption pass the
   budget table's close-shaped range already covers.
4. **CLAUDE.md's gate sentence** was checked: this pass adds no
   tracked Python file, so the sentence's set-by-role Python wording
   (already corrected at an earlier close) stands unchanged. Two
   corrections found in the checking: the `./...` parenthetical still
   named `cmd/dubplate-server` and `internal/config` as "today"'s
   packages, stale since rung 8's own packages joined the module gate
   before this pass even started, so the aside is dropped; and the
   state-dir agreement leg, a `check.sh` leg the sentence did not yet
   name, is added to the listed legs.
5. **`docs/HISTORY.md:442-447`'s forward-carry-5 correction**, which
   the spec's own errata section already named, had not been applied
   at 8a's close; it is applied by this entry, at the rung 4b entry's
   item 5 (`docs/HISTORY.md:1876-1883` as this entry lands; the pointer
   moves again the next time a pass prepends an entry above it, the
   same drift that made the original `:442-447` pointer stale).

### The evidence

The golden's PICTURE block arithmetic: type index 3, media type
`image/png`, description `cover`, geometry all zero, 32 data bytes
beginning with PNG's signature, the decoded field widths summing to the
block's declared length of 78 (checked against the literal 78 after the
leftovers round, not against the block's own recorded length). `go test
./internal/engine/tagio/... -run Conformance -count=1` runs
`TestConformanceLedgerSubjects` with twelve subtests in about 0.05s; the
package's own suite runs in about 0.1s. The twelve subject ids:
`clean-match`, `media-uneven-split`, `va-exact`,
`non-latin-pseudo-release`, `year-reissue-ratio-penalty`,
`real-classical-karajan-beethoven-89`, `real-reissue-nevermind-2011`,
`real-boxset-james-brown-star-time`, `real-nonlatin-utada-first-love`,
`real-noise-dylan-bootleg`, `real-genre-public-enemy-nation-millions`,
`real-decoy-mozart-requiem`. All twelve agree with the ledger on all 69
keys; zero divergences, so nothing is staged into
`corpus/analysis/adjudications.json`. `internal/harness/analysis` gains
its own test coverage for the schema addition
(`adjudication_test.go`), and its existing suite passes unchanged.
`go test ./... -count=1` (every package) is green in 2m28.693s, no
existing suite's result changed by this pass. `bash scripts/check.sh`
is green in 2m20.204s, gitleaks the dominant cost at about 2m14s
scanning 235.58 MB, now also covering lane 2's worktree (Task 0's
about-three-minutes estimate, above); no new tracked file was flagged.

### What a later pass would be wrong to rediscover

The seam registry is the one dispatch point for `Read` and `Write`;
`Parse` is a deliberate, tested bypass, not an oversight (m6), so a
future pass that wants `Parse` to route through the registry is
reopening a ruled decision rather than closing a gap. `Write`'s picture
argument's three states mirror `TagSet`'s own list-member idiom
exactly (nil, non-nil empty, non-empty); a caller that wants to leave
pictures untouched passes nil, never an empty slice built from the
current read. `Picture.Data` aliases the decoded block's own payload
rather than copying it, so a caller mutating it in place mutates what a
later `Write`, called with a nil pictures argument, preserves untouched.
The instrument's 69-key comparison exists because `Write` rewrites or
deletes every owned key on every call (the fact that makes Task 5's
mutation 6 unfirable, above); a future pass adding a partial-write mode
reopens that property and must re-derive whichever mutations depended
on it. `Read` staying tag-only is a stated, ruled choice (R12), not an
oversight rung 9 has to notice on its own; rung 9 picks between a
second read entry point and an exported projection over `Container`,
and either choice retires coverage gap 4 above.

### Deviations

Recorded as package documentation in
`internal/engine/tagio/deviations.go` rather than repeated here in
full: APNG discrimination (R3), the CR2 collision (R2), the zeroed
write geometry (P6), and the block-length trust extended to cover
VORBIS_COMMENT as well as PICTURE (P3), the last naming 8a's own
`vorbisField` as the code the policy already applied unrecorded.

### Spend

About 2.9M against the 3.25M ceiling: plan preparation, 0.87M; Task 0,
0.11M; the chain, 1.51M across 16 agents; the leftovers round, 0.18M;
the batched `code-simplifier` round, 0.13M; conductor turns, about
0.1M. The 2.6M flag was crossed by the leftovers round, after every
task had already landed. Attended interactions: zero planning misses,
zero execution sittings.

## Rung 8, pass 8a (the tagio writer core and reader): closed 2026-09-08

Tasks 0 through 9 executed against
`docs/superpowers/plans/2026-09-07-rung-8a-tagio-writer.md` (revision 2)
in the main tree. This entry is Task 9, the pass close. Task 0 found no
blockers: `.gitleaksignore` is unchanged and empty of new pins, and
`gen_template_flac.py`'s path sat one directory level deeper than the
plan assumed, a drift note rather than a stop.

### What landed

`internal/engine/tagio` is born this pass, plus one added exported
function on `internal/atomicfile`. Task 1 (`149d10c`) added
`atomicfile.WriteStream` (`internal/atomicfile/atomicfile.go:55`), the
streaming atomic write `tagio`'s writer rides: a temp file in the
target's own directory, the caller's write function, sync, chmod to the
caller's mode, rename, then a directory fsync. `WriteFile` is
unchanged. Task 2 (`986eb1c`, fix `81efc0d`) authored
`internal/engine/tagio/testdata/golden.flac` and its provenance record,
`golden.txt`, generated once from the pinned oracle venv (beets 2.13.1,
mediafile 0.17.0, mutagen 1.48.1) over a declared input reaching every
one of Task 8's ledger-silent keys plus the four corpus-unreachable
cases; the fix restored a real second date so the golden still exercises
month/day decomposition, and corrected the input's attribution to name
the covering shape as the generator's own rather than a reproduction of
an `INV` input list, since `INV:409-427` publishes an emitted shape, not
an input.

Task 3 (`f7c8ffc`) added `Parse` (`internal/engine/tagio/container.go:122`),
the container reader: the block chain in order, STREAMINFO and
VORBIS_COMMENT decoded, and the frame region as `Container.FrameOffset`
and `Container.FrameLength` rather than as bytes (C15, C7a). It refuses
all seven of C7's malformed and unsupported shapes, including the two
this pass's review folded and the conductor ruled 2026-09-08: a comment
field name outside U+0020-U+007E, and a second SEEKTABLE block, both no
longer non-refusals under C10's general preserve-unknown-by-length rule
(RFC 9639 section 8.5 allows at most one). Task 4 (`8c0722d`, fix
`c6ea3f6`) added `TagSet` (`internal/engine/tagio/tagset.go:115`) with
its four states carried in Go-native pointer and slice types, and
`Mapping()`/`WriteOrder()` (`:332`, `:320`) reproducing `INV:315-385`'s
69-row table with the three colliding pairs resolved plural-last; the
fix repaired two subtests that built assertions on bare maps instead of
reading `TagSet`'s own fields, so a dropped read never actually failed
them.

Task 5 (`8c03957`, fix `a12d81c`) added `Serialize`
(`internal/engine/tagio/encode.go:33`) and the `safe_cast` read-side
primitives, covering every serialization rule in the package contract:
float places, the `" dB"` suffix, the R128 Q8 shift, list repetition,
the bool encoding, and date decomposition with its four-digit year and
unpadded `YEAR` fallback. The fix replaced a naive per-invalid-byte
`utf8.DecodeRuneInString` scan with `utf8SequenceSpan`
(`internal/engine/tagio/encode.go:460`), which collapses a maximal
ill-formed subpart into one U+FFFD the way Python's `errors='replace'`
does, a real divergence the fix's four table rows pin. Task 6
(`933f043`, reviewed clean by opus) added `Read`
(`internal/engine/tagio/read.go:14`), the exported 69-key read
projection: first-truthy-style precedence, C5's last-style rule,
`safe_cast`'s numeric behavior, date recomposition, and the Q8
division, asserted in one equality against golden.txt's transcribed 69
values.

Task 7 (`88e7ecf`, fix `4834199`) added `Write`
(`internal/engine/tagio/write.go:46`), the merge contract: a
case-insensitive delete-then-append over `Serialize`'s fixed order,
every unowned block preserved under a re-emitted header, the vendor
preserved on rewrite and freshly named (`dubplate `-prefixed, per R9)
on creation, exactly one 8 KiB PADDING block written last, and the
frame region copied verbatim. It refuses a malformed or unsupported
container before the target is ever opened for writing. The fix widened
the surviving-comments assertion from one colliding pair to all three.
Task 8 (`b42e3b2`) added the ruling 1 unit tests over the sixteen
ledger-silent keys plus `bpm`, and the six corpus-unreachable cases
(zero year, empty list, `catalognum`/`barcode` empty, invalid UTF-8, the
U+007E field name), and `deviations.go`, the package-documentation
deviation note rung 15's launch gate reads directly.

The exported surface, as it stands at close: `Parse` and `Read` for the
two read paths, `Write` for the merge contract, `TagSet` with its four
states, `Mapping()` and `WriteOrder()` for the 69-key table and its
fixed write order, `Serialize` for the encoding rules, and the two
sentinels `ErrUnsupportedContainer` and `ErrMalformedContainer`
(`internal/engine/tagio/container.go:41`, `:46`). No signature carries a
picture parameter (R5); 8b widens `Container` and `Write`'s argument
list for PICTURE as one in-package refactor with no consumer wired yet.

The leftovers round (`8ff65bd`) closed six findings across the accepted
tasks in one post-chain pass, without a second reviewer (below): the
count-and-names test enumerated `Mapping`'s rows rather than `TagSet`'s
members, so a dropped struct member could not fail it, closed by
`TestTagSetMembersTieToMapping`; `Mapping` was unexported and returned
its backing slice, now `Mapping()` returns a `slices.Clone`, proven by
`TestMappingAccessorReturnsCopy`; `emitBlock`'s payload length overflows
its three-byte field at 1<<24 bytes, closed by `checkBlockPayloadSize`
and `TestCheckBlockPayloadSize`; the U+007E field-name test asserted
only that `TITLE` still projected, not that the tilde-named comment
itself survived, tightened; three comment and doc fixes (a garbled
sentence, a nonexistent stdlib name, one wrong noun); and CLAUDE.md's
gate sentence, which named a stale Python file count, rewritten to name
the files by role.

### What the gate caught

Task 4's fix (`c6ea3f6`) is a mutation-testing catch turned real defect:
the four-state table's list and numeric subtests built `[]string` and
`*int` values directly rather than reading them off a constructed
`TagSet`, so `TagSet.Genres` and `TagSet.Track` were named in the
failure message but never actually read; the second `diff-reviewer` pass
caught it and the conductor folded the fix in the leftovers round rather
than a third chain round. Task 5's fix (`a12d81c`) is a real behavioral
divergence the plan's own citation work found, not a review nit: mutagen
decodes invalid UTF-8 with Python's maximal-ill-formed-subpart rule, and
a naive per-byte scan emits one U+FFFD per invalid byte instead of one
per subpart. Task 7's twelve named mutations ran and restored: ten fired
as the plan predicted; one (opening the target before validating the
container) needed a dedicated test to observe, since
`atomicfile.WriteStream`'s own cleanup on error masked the reordering
from a bytes-unchanged assertion alone, closed by adding
`TestWriteRefusalNeverCreatesTempFile`; one (a whole-file read on the
write path) stayed unfired by design, held by the reviewer's read per
the plan. Every other task's fix round is a targeted correction the
`diff-reviewer` sent back once; none needed a second re-dispatch.

### The resolutions and rulings, as applied

All ten resolutions stand as ruled, conductor 2026-09-08, per the plan's
Resolutions section; none folded.

- **R1** (the golden's generator is a committed, hand-run script beside
  the golden): `internal/engine/tagio/testdata/gen_golden_flac.py`.
- **R2** (one golden, widened to reach every ledger-silent key and every
  corpus-unreachable case): `internal/engine/tagio/testdata/golden.txt`'s
  declared-input section.
- **R3** (the oracle template is read at generation time only, never by
  a Go test): `gen_golden_flac.py`'s copy step; no `_test.go` file in
  the package opens a path outside `internal/engine/tagio`.
- **R4** (the byte assertion writes into a copy of the golden, not the
  template): `write_test.go:115` (`TestWriteGoldenBytes`).
- **R5** (no picture parameter on 8a's signatures): `write.go:46`
  (`Write`'s argument list), `container.go:102` (`Container`).
- **R6** (no `ctx` on `tagio`'s entry points): `read.go:14`, `write.go:46`.
- **R7** (both sentinels land in 8a; 8b adopts the unsupported one
  unchanged): `container.go:41`, `:46`; crosses `spec:1103`, corrected
  below.
- **R8** (the four-state representation is the implementer's, under two
  stated properties): `tagset.go:115`, proven by
  `tagset_test.go:323` (`TestFourStates`).
- **R9** (the fresh vendor asserted by prefix, version from the build):
  `write.go:184` (`freshVendor`), `write_test.go:481`.
- **R10** (`WriteFile` untouched; the streaming write is its sibling):
  `internal/atomicfile/atomicfile.go:55`.

**C7's sixth refusal (double SEEKTABLE) is reversed from the draft
reading**, ruled by the conductor 2026-09-08: RFC 9639 section 8.5
allows at most one SEEKTABLE block, so `tagio` refuses a second rather
than copying it through under C10's general rule. **C9a's fresh-block
placement is confirmed, not corrected**: `golden.txt`'s recorded block
chain shows VORBIS_COMMENT immediately after STREAMINFO, matching the
oracle measurement, and `TestWriteNoCommentBlockInsertsAfterStreamInfo`
grades the writer to it. **The date triple's compose-as-a-unit rule**
(an absent year deletes `DATE` and its `YEAR` fallback together; an
absent month or day beside a present year is treated as falsy) is
applied at `encode.go:330` (`serializeDate`) and proven by
`encode_test.go:73` (`TestSerializeDateTable`) and `:118`
(`TestSerializeOriginalDateEveryComponentZero`). **The 4.5M ceiling**
held; the pass did not cross it (Spend, below).

### The mutation table

Every task's named mutations, as the plan predicted them and as the
implementer reported running them; per-mutation counts below are the
plan's own list, and outcomes beyond a task's commit message are
attributed to the implementer's report rather than restated in detail
here.

**Task 1** (5 named): the mode-drop and early-return mutations fire on
the mode and leftover-temp rows; the directory-fsync drop and the
buffered-internal-`WriteFile` mutation are both reported unfired by
design (a power-cut property and a memory property, neither observable
from a workstation test; both held by the doc comment, ruling 12, and
the reviewer's read).

**Task 2** (2 named): the regenerate-with-one-field-changed mutation
fires the recorded-sequence check; the one-byte-truncation mutation is
reported unfired by design, since Task 3 is the first task that parses
the golden and carries the matching mutation itself.

**Task 3** (11 named): reported by the implementer as run per the
plan's predictions, covering the header-endianness asymmetry, the
STREAMINFO offset window, all seven refusal rows including the two this
pass added, repeated-key merging, and the U+007E preservation pair. The
whole-file-read mutation is reported unfired by design, held by the
reviewer's read for C7a.

**Task 4** (6 named): the field-drop, colliding-pair-order, and
determinism mutations fire on `TestFieldCountAndNames`,
`TestCollidingPairsWriteOrder`, and `TestWriteOrderDeterministic`; the
two empty-state-collapse mutations are kept and graded separately per
`TestFourStates`, which the ledger's 45 empty-list-versus-149
empty-string split argues for; the read-only-mark-drop mutation on
`barcode`'s `UPC` style is reported unfired in this task by design,
since its consequence is a write Task 7 grades.

**Task 5** (9 named): reported by the implementer as run per the plan's
predictions, covering float formatting, the `" dB"` suffix asymmetry
between the gain and peak rows, the R128 Q8 shift, year padding, the
zero-month `-DD` guard, the empty-string-versus-empty-list write
divergence, invalid-UTF-8 replacement, and the compose-as-a-unit
`DATE` rule. One (stripping the suffix after rather than before the
cast) is reported unfired by construction: `safe_cast`'s leading-prefix
match makes the ordering unobservable across all 69 keys, cited against
`mediafile/utils/type_conversion.py:52`.

**Task 6** (7 named): reported by the implementer as run per the plan's
predictions, covering style-break precedence, the last-style rule on
`catalognum`/`barcode`, list projection, the R128 divisor, absent-key
typed zeros, and a two-field-swap caught only by the golden's whole-set
comparison. One (stripping the suffix after the cast) is reported
unfired by construction, the same citation as Task 5's mirrored case.

**Task 7** (12 named), the pass's most detailed report: ten fired as
predicted (big-endian header fields, last-style write coverage, the
foreign-key merge, the padding-and-append guard, the last-block-flag
placement, STREAMINFO's untouched MD5, the frame-region hash, the
write-target-unchanged refusal guard, the caller-mode chmod, and
map-order iteration); one (opening the target before validating)
required `TestWriteRefusalNeverCreatesTempFile` as a dedicated test,
since `atomicfile.WriteStream`'s own error-path cleanup otherwise
masked the reordering; one (a whole-file read on the write path) stayed
unfired by design, the reviewer's read the acceptance criterion names.

**Task 8** (4 named): the computed-not-hard-coded mutation is a review
check rather than a test run, and the reviewer rejected it as the
self-derived-expectation anti-pattern it is; the `" dB"`-suffix-drop
mutation fires here too, the deliberate redundancy with Task 5; the
`BPM`-as-absent-at-zero mutation fires the `bpm` row. The
remove-a-key-from-the-deviation-note mutation is reported unfired by
design: the note's completeness is a reviewer's read against the task's
key list, not a test assertion.

### The reviewer-read proofs

Two properties in the package contract are held by a reviewer's read
rather than by an assertion, and this close records who verified each.
**No whole-file read in `Parse`**: `container.go:122`'s doc comment
states the parse reads the metadata region only and bounds-checks every
declared length before a slice is taken; Task 3's `diff-reviewer` pass
read the diff against this criterion and accepted it, and the leftovers
round's mutation-testing catch on `parseBlocks`'s short-read handling
(`8ff65bd`, item 7) touched the same code path without finding a
buffered whole-file read. **The frame region copied, not buffered, in
`Write`**: `write.go:252`'s `copyFrameRegion` streams from the source
file's offset directly into the atomic-write pipe; Task 7's
`diff-reviewer` pass accepted the diff against this criterion on the
first round, and the twelve named mutations' eleventh (reading the
source whole before rewriting) is the one Task 7 itself reports unfired
by design, naming the same reviewer's read as the proof.

### The declared coverage gaps

Seven gaps this pass builds or uncovers and no test in this pass can
reach, joined to the running total of thirty-six (rung 3's twenty-one,
rung 4a's nine, rung 4b's four, rung 1b's two, `HISTORY:241-259`),
bringing the total to **forty-three**. A container carrying two
SEEKTABLE blocks is not on this list: ruled a refusal rather than a
non-refusal (above), Task 3's `TestParseRefusals` and `TestWriteRefusals`
grade it, so it never became a gap.

1. **The directory fsync's crash property is unproven.** Task 1's
   mutation 2 names it; held by the doc comment on
   `atomicfile.WriteStream` and ruling 12, not by a test, because a
   power-cut is not observable from a workstation test.
2. **The streaming property is held by a reviewer's read, not an
   assertion, at three sites**: Task 1's buffered-write mutation, Task
   3's whole-file-read mutation (C7a), and Task 7's whole-file read on
   the write path. All three hold one property, that a multi-gigabyte
   rewrite costs a fixed amount of memory rather than a file-sized one.
3. **`atomicfile`'s cross-filesystem rename behavior is untested** on a
   single-filesystem workstation, inherited from rung 4a's own gap
   list.
4. **No test grades against the twelve ledger subjects.** That is 8b's
   whole pass; 8a's read-back agreement rests on the golden and on
   ruling 1's unit tests until 8b's conformance instrument runs.
5. **PICTURE is parsed and preserved, never decoded.** 8b owns its
   fields and its geometry rule.
6. **No real library file has been written.** Every fixture in this
   pass is the golden, a `buildContainer` product, or a writer product;
   a real beets-written library file is first written at rung 9.
7. **A genuine mid-chain I/O error is indistinguishable from EOF in
   tests.** The leftovers round's fix to `parseBlocks` (`8ff65bd`, item
   7) now propagates any `ReadAt` error that is not `io.EOF`, but
   `Parse` takes only a path and opens its own `*os.File`, and a
   regular file's short `ReadAt` past the declared block chain always
   reports `io.EOF`, so there is no cheap device on this workstation to
   inject a failing `io.ReaderAt` without changing `Parse`'s signature.
   An injectable reader is 8b's or 8c's call.

### The measured evidence

`golden.flac` is 2222 bytes: STREAMINFO (34 bytes), VORBIS_COMMENT
(1066 bytes), PICTURE (78 bytes), PADDING (1024 bytes, last), each with
its 4-byte header, matching the file size exactly (a silent template
with no audio frames). **C9a**: the golden's VORBIS_COMMENT sits
immediately after STREAMINFO, confirming the fresh-block-placement
hypothesis; no correction to the plan followed. **The comment-order
hypothesis**: the golden's recorded on-disk sequence matches the
plan's prediction (`sorted_fields()` field order, then style
declaration order within a field, then list order within a style) with
no mismatch to report, including the `albumtype`/`albumtypes`
collision, which resolves the plural in place rather than moving its
position. `go test ./internal/engine/tagio/... -count=1` runs 59 tests
in 0.121s within the full-suite run; `go test ./... -count=1` (every
package) is green in 2m20s with no existing suite's result changed,
since this pass touches no landed package but `internal/atomicfile`.
`bash scripts/check.sh` is green in 1m26s, gitleaks the dominant cost at
1m19s scanning the tracked worktree (117.85 MB); `golden.flac` was not
flagged, and `.gitleaksignore` needed no new fingerprint.

### What a later pass would be wrong to rediscover

`Mapping()`'s ordering is the write order and the only order: a caller
that wants the colliding pairs resolved plural-last iterates
`WriteOrder()`, and `Mapping()` itself returns a `slices.Clone`, so
mutating a caller's copy never touches the package's table. `Parse`
takes a path rather than an `io.ReaderAt`, which is why a genuine
mid-chain I/O error cannot be distinguished from a clean end of chain in
this package's own tests (gap 7); a future pass that wants that
distinguished needs to widen `Parse`'s signature first, not add a test
against the current one. The invalid-UTF-8 replacement follows the
Unicode maximal-ill-formed-subpart rule, not a naive per-byte scan, and
`utf8SequenceSpan` is where that lives; a rewrite that reaches for
`utf8.DecodeRuneInString` per byte reintroduces the exact divergence
Task 5's fix closed. `checkBlockPayloadSize`'s boundary is the
three-byte length field's own limit (1<<24), not a memory ceiling
someone picked; grade it on the length alone, never with a 16 MiB
allocation. The date triple composes as a unit: an absent month or day
beside a present year is falsy, not a reason to omit `DATE`, and this
diverges from mediafile's own read-modify-write semantics by design,
recorded in `deviations.go` rather than rediscovered as a bug.

### The 8a to 8b handoff

8b calls `tagio.Read` as the package's one read entry point rather than
carrying a second projection; its conformance instrument grades the
twelve ledger subjects through it directly. 8b regenerates no golden:
`testdata/golden.flac` already carries a PICTURE block, and `golden.txt`
records its decoded header, so 8b's byte assertion reads the same file
8a committed. The refusal set (`container.go:41`, `:46`, and
`TestParseRefusals`'s seven rows) is closed; 8b adds no new refusal
without a fresh ruling. `ErrUnsupportedContainer` is landed and 8b's
seam registry adopts it unchanged rather than reclassifying an error
that already exists, per R7's crossing of `spec:1103` (corrected,
below): the registry, its func-value dispatch, and its supported set
are 8b's, and 8b's own sentinel work is only the malformed-versus-
unsupported split, already done. The PADDING policy (drop every
incoming block, write exactly one 8 KiB block last) is `write.go`'s and
does not change under PICTURE. `Mapping()` is the exported accessor for
8b's own citation work if it needs the 69-row table; it returns a copy,
never the package's backing slice.

### Forward carries

1. **`code-simplifier` did not run over 8a's Go** and is owed at 8b's
   opening, per the conductor's window-cap ruling (below).
2. **The `io.ReaderAt` gap** (coverage gap 7, above) is 8b's or 8c's
   call: whether `Parse` grows a second, injectable entry point or
   stays path-only.
3. **Fork B stays open and owed to Geoff**, unchanged by this pass: a
   FLAC-only 1.0 with the MP3 pass immediately after cutover, applied
   provisionally by the conductor per the plan's header.
4. **The group non-FLAC question** (whether a member's non-FLAC
   library gets a rung 8-shaped pass or stays out of scope) is
   unresolved and carried, not decided here.

### The deviations recorded

Two deviations from Task 9's own review, both already package
documentation in `internal/engine/tagio/deviations.go` rather than
repeated here in full: the date triple's compose-as-a-unit rule
diverges from `DateItemField.__set__`'s read-modify-write semantics
where `tagio`'s write entry point has no partial-update surface to
express the difference, and `tagio`'s field-name bound is RFC 9639's,
one code point wider than mutagen's `is_valid_key` bound, unreachable
at the encoder because C1's constants and C7's refusal make an invalid
name impossible to construct on the write side.

### Two deviations from the plan's own execution model

Task 7's conditional opus upshift (triggered by Task 5's fix verdict,
per the plan's escalation rule) was not applied, by conductor ruling
under the unattended-window cap, with the `diff-reviewer` standing as
the net; Task 7 landed clean on its own review round regardless. The
leftovers round (`8ff65bd`) ran without a second reviewer, for the same
window-cap reason; its six findings are individually small and each
carries its own proof (above) rather than a fresh adversarial pass.
`code-simplifier` owed at 8b's opening (forward carry 1) is the same
cap's third consequence.

### Spend

About 3.6M against the 4.5M ceiling: plan preparation, 0.87M; Task 0,
0.1M; the chain, 2.27M across 24 agents; the leftovers round, 0.2M;
conductor turns, about 0.15M. The 3.6M flag was not crossed. Attended
interactions: one planning miss (Task 4's count-and-names criterion let
the test enumerate the mapping rather than the members, closed in the
leftovers round by `TestTagSetMembersTieToMapping`), zero execution
sittings.

## Rung 1b (Phase 0b, the config wall-row registry): closed 2026-09-08

Tasks 1 through 6 executed against
`docs/superpowers/plans/2026-09-07-rung-1b-config-registry.md`
(revision 2.1) in the main tree. This entry is Task 7, the pass close.

### What landed

Task 1 (`3a35558`) retired the `.gitleaksignore` line pin: renamed the
fixture field the generic-api-key rule keyed on, in both tables of
`internal/config/keys_test.go`, rather than re-fingerprinting a file
that would only go stale again. Task 2 (`060f743`, fix `859f26c`) added
`Registry`, the collision-gate panic, and its six O5 checks; the fix
closed a within-class repeat hole the first pass left open (a key
repeated inside one `ReservedClass` or `IntegrationClass` literal, or
listed in both `Required` and `Optional` of one integration class,
registered silently). Task 3 (`82169b6`) migrated rung 1's own
declarations onto the registry and rewired `Load` and doctor to take it
as a parameter; three citations left dangling by `Register`'s deletion
were retargeted in the post-chain round. Task 4 (`87053c3`) absorbed the
integration family as data and retired `Load`'s `class.Name == "shadow"`
branch; rung 4's shadow class and its two thresholds moved to
`cmd/dubplate-server/shadow.go`, and `TestLoadShadowIntegration`'s eight
rows moved with it, verbatim, to `shadow_test.go`. Task 5 (`ae1e951`)
added `KindFloat` and `KindString` and moved threshold validation to
registration time, parsing each declaration's own `Default` under its
declared `Kind`; a doctor test crash (four `t.Errorf` calls followed by
an unguarded slice index) was fixed separately (`1ab8c93`) before Task 6
landed. Task 6 (`1952ea8`) derived doctor's line count from the registry
and stated Resolution R10's both directions, the classification
completeness the shipped template satisfies and the greenfield
template's owned-row completeness it does not yet.

The post-chain round (`b7e9d71`) closed four gaps the chain left open:
the prefix check had to reach the thresholds family too, which forced
renaming a threshold test fixture off its own unprefixed literal
(`TEST_STRING` to `DUBPLATE_TEST_STRING`); `AddIntegrationClass` did not
claim `Optional` keys across families, so an Optional key colliding with
an owned or reserved name registered silently; Task 2's `Flag`-specific
subtests asserted the generic duplicate message rather than the
`Flag`-naming one; and three more citations had gone stale under the
same deletion. Its own `diff-reviewer` sent back a fix round (`e364aa4`):
hoisted the prefix check out of `claim` into a standalone
`requirePrefix`, called unconditionally before any claim or claim-skip
decision, so the R5 Optional-plus-Threshold pairing's claim skip could
never bypass it; added a reverse-order test for that pairing; and
corrected one more stale citation and a doctor test's kind-name
assertion. `code-simplifier` (`d5c44ac`) then extracted `panicDuplicate`
for the eight repeated collision messages and a generic `parseThreshold`
helper shared by `ParseInt`, `ParseDuration`, and the new `ParseFloat`.

### What the gate caught

Task 2's within-class repeat hole (above) drew a second `diff-reviewer`
pass before it was accepted. Task 3 left three `thresholds.go:` and
`registry.go:` citations dangling after `Register`'s deletion, closed in
the post-chain round. Task 5's registration-time check carried a carve-out
for thresholds that skipped the `DUBPLATE_` prefix rule, on the reasoning
that a threshold is read directly by `wall.get` and never classified by
`rejectNamespace`'s prefix walk; the conductor reversed the carve-out in
the post-chain round, forced by the plan's own fixture: a threshold test
declared under the literal `TEST_STRING`, with no prefix, was the exact
case the reversed rule now refuses at registration rather than letting
register unprefixed and unreachable. Task 4's Optional keys went
unclaimed cross-family until the round closed it, proven by
`TestCollisionGate_OptionalCrossFamily`. The round's own four evidence
gaps (the missing reverse-order pairing test, an uncomputed citation, an
underspecified panic-message assertion, and a doctor test's literal kind
name) were all closed in the fix round `e364aa4`.

### The contract change, recorded as a contract change

Rung 1's config-core contract stated two things: every `DUBPLATE_*` name
falls into one of five classes, and `[Thresholds]` is a separate,
open-ended registry distinct from the wall, which was a closed table.
**After rung 1b, the wall stops being a closed table and becomes
registry contents.** The five-class classification stays the contract;
thresholds are now one declaration family among four (owned wall rows,
reserved classes, integration classes, thresholds), all registered
through the same `Registry` value. `Load` takes the registry it enforces
against rather than reading package-level tables. A rung declares its
keys in its own file, inside the package that owns them, and no lane
adds a name to `internal/config/keys.go`; a lane removes only its own
reserved class from that file, in the same commit that declares the
class in its own package (rung 5 is first, taking `auth and identity`
and `serve`). The build spec's `:182-183` sentence and the parallel
spec's `PA:56` paragraph are corrected to match, both below.

### The sixteen resolutions, ruled

All sixteen are ruled as applied, 2026-09-07, per the plan's revision 2
header; two markers carry 2026-09-08, the date the post-chain round
reversed or hardened them. Location is where the application sits at
HEAD, drawn from the task returns and this pass's close commit
(`db040ec`).

- **R1** (the registry is a value `cmd/dubplate-server` assembles and
  `Load` takes): `internal/config/config.go:199`, `:240`;
  `internal/config/keys.go:35`, `:153`, `:192`;
  `internal/config/registry.go:33`; `cmd/dubplate-server/shadow.go:109`;
  `cmd/dubplate-server/main.go:62`.
- **R2** (the two class shapes, `ReservedClass` and `IntegrationClass`,
  stay distinct): `internal/config/registry.go:210`.
- **R3** (a collision panics at registration; a gate-resident test
  constructs the production registry): no single inline marker, the
  ruling covers the panic sites `panicDuplicate` serves throughout
  `internal/config/registry.go` plus the eleven-test
  `TestCollisionGate_*` family at `internal/config/registry_test.go:97-336`.
- **R4** (rung 1's own rows stay declared inside `internal/config`):
  `internal/config/registry.go:63`.
- **R5** (`Load`'s `class.Name == "shadow"` branch retires; the two
  class shapes are re-expressed as a pairing exemption): general
  application at `internal/config/keys.go:225`,
  `cmd/dubplate-server/shadow.go:27`; the pairing exemption itself,
  ruled 2026-09-08 in the post-chain round, is the code at
  `internal/config/registry.go:244-249` (`AddIntegrationClass`'s skip)
  and `:279-291` (`isOptionalIntegrationKey`), guarded by
  `TestCollisionGate_OptionalCrossFamily` and
  `TestCollisionGate_OptionalThresholdPairing_ReverseOrder`.
- **R6** (`KindFloat`'s mechanism lands here; the four knobs do not): no
  code surface this pass; carried forward, below.
- **R7** (`AdminEmail`'s dedicated contact row does not land here): no
  code surface this pass; carried forward, below.
- **R8** (the superseded map is 1b's by necessity): `internal/config/registry.go:340`.
- **R9** (the gitleaks fix is the field rename plus a comment, not a
  `.gitleaks.toml`): `internal/config/keys_test.go:59-64`, applied
  2026-09-07 at `3a35558`.
- **R10** (1b adds no name to any template; the coupling gets a test and
  a stated rule): `internal/config/config_test.go:226`, `:262`;
  `cmd/dubplate-server/doctor_test.go:150`.
- **R11** (`KindInt` stays the zero value; an unnamed or mis-defaulted
  kind is a registration error): `internal/config/registry.go:266`,
  `:305`; `internal/config/thresholds.go:13`.
- **R12** (pattern key families and `DUBPLATE_COLLECTIONS` declined and
  filed to rung 9): no code surface this pass; filed, not carried, since
  no rung yet owns the class.
- **R13** (the three missing reserved classes recorded, not declared):
  no code surface this pass; filed to rungs 5 and 12's own spec work.
- **R14** (rung 4's ruling 1 re-expressed in registry terms is 1b's own
  migration work): landed as Task 4, same locations as R5's general
  application above.
- **R15** (declaration order is preserved; the registry never sorts):
  `internal/config/registry.go:185`.
- **R16** (every declared key carries the `DUBPLATE_` prefix; the
  pending row is not a key): general carve-out ruled 2026-09-07 at
  `internal/config/registry.go:137`; the threshold-family clarification,
  that the prefix check reaches thresholds too and cannot be bypassed by
  `AddThreshold`'s R5-pairing skip, ruled 2026-09-08 in the post-chain
  round and hardened in the fix round, at `requirePrefix`
  (`internal/config/registry.go:106-110`), called unconditionally from
  every `Add*` entry point.

### The mutation table

Every named mutation Tasks 1 through 6 report, reported by the
implementer and re-confirmed for Task 2 in the post-chain round
(`b7e9d71`, "the seven named Task 2 mutations re-ran clean against the
current tree, mutation 1 confirmed 4 of 4 families, mutation 5 confirmed
independently").

**Task 1** (2 mutations): 1, revert the rename → gitleaks reports one
finding again, fired. 2, add a high-entropy row to each table in turn →
zero findings both times against the fix as landed; against a rename
scoped to `TestReservedClassOf` alone, `TestSupersessionOf`'s row
reported one finding, measured, which is what the fix's second table
closes.

**Task 2** (7 mutations): 1, drop one family's within-family duplicate
check → that family's test fails, confirmed 4 of 4 families. 2, drop the
cross-family check → the cross-family test fails, within-family tests
still pass. 3, drop the prefix check → the prefix test fails. 4, return
the backing slice instead of a copy → the copy test fails. 5, share one
registry across two constructions → the two-independent-registries test
fails, confirmed independently, the proof `sync.OnceFunc`'s retirement
rests on. 6, let a status row through the prefix check → the
identity-source carve-out test fails. 7, drop the integration `Flag`
check → both the empty-`Flag` and `Flag`-in-`Required` tests fail,
reworded in the round to assert the `Flag`-specific message rather than
the generic duplicate one.

**Task 3** (5 mutations): 1, sort reserved classes by rung → the R15
order assertion in `TestDoctorCmd_LiteralSpotCheck` fails, `TestDoctorCmd`
itself stays order-blind and passes, both outcomes reported. 2, drop the
identity-source `Pending` handling → the literal-spot-check fails and a
satisfied wall starts requiring a key named `identity source`. 3, drop
the rung number from a reserved-class error → the contract table's
"reserved key early" row fails. 4, route an accessor around `wall.get`
directly → **reported as a gap**: no existing test covered the bypass
for the accessor chosen, so the mutation could not be run against a
failing test; folded into this pass's own coverage-gap count, below. 5,
register rung 4's thresholds onto one registry and pass a fresh second
one to `newDoctorCmd` → `TestDoctorCmd_ShadowRows` fails on both
threshold lines, the proof the wiring is a value pass rather than a
hidden global.

**Task 4** (7 mutations): 1, resolve every class's values regardless of
enabled state → the flag-unset row fails, proving the reworded
assertions still fire. 2, accept every integration key unconditionally
in `rejectNamespace` → the flag-unset row fails. 3, move both threshold
names from `Optional` to `Required` → the flag-set-with-every-path row
fails, reporting both missing. 4, read an unparseable flag as off → the
unparseable-flag row fails. 5, drop the `name == class.Flag` guard → both
explicit-false rows fail. 6, restore a class-name branch with a second,
branchless class → the second class's resolved values come back empty,
the proof outcome 1 (the branch's retirement) landed. 7, derive enabled
state from whether any key resolved → the empty-`Required` row fails,
the hook 6 shape.

**Task 5** (5 mutations): 1, make the float parser's bad-default and
bad-value messages match → the fourth test row fails, the two are
asserted to differ. 2, drop default validation, keep enum-membership →
the float-defaulted threshold with no explicit `Kind` registers clean
and doctor prints `invalid: invalid default "0.04"`; both halves
reported, the proof the enum check alone hides the defect. 3, drop the
kind check from the collision gate entirely → the registration test
fails, both the unnamed-kind and unparseable-default thresholds register
silently. 4, change `KindFloat`'s value so `KindInt` is no longer the
zero value → every `Threshold` literal that omits `Kind` changes
meaning, existing integer threshold tests fail, the proof R11's
zero-value choice is load-bearing. 5, drop `KindString`'s empty-default
check → the scratch threshold with an empty `Default` registers clean
and later reads as unset rather than as its declared default.

**Task 6** (5 mutations): 1, remove the new declaration's family from
the scratch registry's accessor → the scratch test fails while the
landed `doctor_test.go:127-129` count assertion still passes, both
outcomes reported, the pair showing the new test covers ground the old
one does not. 2, remove one family's block from doctor entirely → the
landed count assertion fails, naming the group by its count
difference. 3, point the classification test at a file assigning no
`DUBPLATE_*` names → the zero-names guard fires. 4, point it at a
nonexistent path → fails at the open, proving a moved template is fatal
rather than silently skipped. 5, remove `DUBPLATE_LIBRARY_ROOT` from the
greenfield template the completeness test reads → the completeness test
fails naming that row; reverted without committing the template edit.

### The declared coverage gaps

Two gaps this pass builds and no test in this pass can reach, joined to
the running total of thirty-four (rung 3's twenty-one, rung 4a's nine,
rung 4b's four, `HISTORY:158-163`), bringing the total to **thirty-six**.

- **The registry's collision panic is unreachable in the deployed
  binary as shipped.** The gate-resident `TestCollisionGate_*` family
  catches every collision before a commit lands, so the panic path
  itself is proven only by tests that recover it, by design.
- **`wall.get`'s choke-point property has no mechanical guard.** Task
  3's mutation 4 (routing an accessor around `wall.get` directly) found
  no existing test covering the empty-value-equals-unset property for
  an accessor that bypassed it; a future accessor written directly
  against the environment map would pass every other test.

Two further items are declined and filed rather than counted as gaps,
since neither is a surface 1b itself builds: pattern key families and
`DUBPLATE_COLLECTIONS` (Resolution R12, filed to rung 9, which owns the
monitoring class and the pipeline), and the three reserved classes the
build spec names and the tree lacks, an SSO proxy, a trusted-proxy mode,
and a forward-auth front (Resolution R13, filed to rungs 5 and 12).

### What a later pass would be wrong to rediscover

The R5 pairing exemption is order-symmetric: an `Optional` key may be
registered as a `Threshold` in either order, `Threshold` first or
`IntegrationClass` first, and both directions are gated
(`TestCollisionGate_OptionalCrossFamily`,
`TestCollisionGate_OptionalThresholdPairing_ReverseOrder`). The prefix
check is its own unconditional step, `requirePrefix`, called before any
claim or claim-skip decision in every `Add*` entry point; a check folded
back into `claim` alone will silently exempt whatever skips `claim`, the
exact bug the post-chain round found and the fix round proved with a
table-driven mutation. The gitleaks trigger is a single field name on a
single line, `key` beside a high-entropy `DUBPLATE_*` value in exactly
one production name (`DUBPLATE_R2_BUCKET`); a rename generalizes across
a file, not a table, and the second table in `keys_test.go` is where
revision 1's fix would have missed it. `KindInt` is deliberately the
zero value of `ThresholdKind`, so a threshold declared before a `Kind`
field existed keeps its old meaning; the hazard it creates is caught by
parsing each declaration's own `Default` under its declared `Kind` at
registration, not by an enum-membership check, which a zero-value kind
passes trivially. `parseThreshold` is the one generic parser
`ParseInt`, `ParseDuration`, and `ParseFloat` share the
unset/set-value/bad-default shape through; a fourth kind extends it
rather than triplicating it again. The template classification's fourth
branch, `integrationClassOf`, is mandatory, not optional: the shipped
template carries zero shadow names today, but pass 4b's own template
carries six, and a walk that tries only `ownedNow`, `reservedClassOf`,
and `supersessionOf` reports six names unknown and reds the gate for
every lane.

### Forward carries

1. **The four-knob wall registration is unblocked** and is the named
   task for the first rung 3 pass that runs. `KindFloat` and its parser
   are landed, so that pass registers `StrongRecThresh`,
   `MediumRecThresh`, `RecGapThresh`, and `SearchLimit` and adds no
   registry capability. It must decide whether
   `internal/engine/matcher` may import `internal/config`, which
   `tunables.go:50-54` currently declines in writing, or whether the
   knobs are declared from `cmd/dubplate-server` instead.
2. **`AdminEmail`'s dedicated contact row** stays post-1b, and is now
   cheap: one declaration in whichever package owns the User-Agent,
   plus the owner touchpoint that minting a new required key implies.
3. **How a rung declares its keys after 1b**: in its own file, inside
   the package that owns them, per O1/O2; rung 4's shadow class
   (`cmd/dubplate-server/shadow.go`) is the worked example. A rung
   taking over a reserved class deletes it from
   `internal/config/keys.go` in the same commit that declares it in its
   own package, the registration-time class-arrives-intact check is
   what stops the move from dropping a key, and that deletion is the
   only edit any lane makes to `keys.go` after 1b. Rung 5 is first,
   taking `auth and identity` and `serve`.
4. **Rung 12's streamer class moves to the driver's own package** when
   that package is born, and the three Navidrome names go with it. This
   is hook 6; 1b landed the seam, not the move.
5. **A string or enumerated threshold kind** is the next kind anyone
   will want, for hook 12's forbidden-character set and hook 14's ID3
   version; the mechanism is now three edits and its registration-time
   check is already in place.
6. **The gitleaks pin is gone and is not coming back.** Any future
   declaration file that writes a key literal beside a `Key:` field
   applies the same rename discipline.
7. **Lane-2 is now unblocked.** 0c landed at `b77e159`; 1b's own close is
   what unblocks lane-2 outright.

### Spend

About 3.7M against the 2.5M ceiling: plan preparation, 1.07M; Task 0,
0.14M; the revision 2.1 re-point, 0.24M; the chain, 1.55M across 16
agents; the post-chain round, 0.25M; its review, 0.09M; the simplify
pass, 0.11M; the fix round, 0.13M; conductor turns, about 0.15M. The
2.0M flag was crossed inside the chain and the conductor did not act on
it, a conductor miss this entry records rather than omits. Attended
interactions: one planning miss (the unprefixed threshold fixture,
`TEST_STRING`, that forced R16's threshold-family reversal rather than
catching it in review), zero execution sittings.

## Rung 4, pass 4b (closed 2026-09-08 at the sitting)

Tasks 0 through 7 executed in the main tree against
`docs/superpowers/plans/2026-09-07-rung-4b-deploy.md` (revision 2) on
2026-09-07. Task 8 handed off to Task 9, Geoff's attended deploy sitting,
which ran 2026-09-08 and closes the pass below.

### What landed

Task 0, the staleness pre-task, ran clean with no blockers. The chain
landed six tasks: Task 2 (`c8b6d71`, fix `30f3faa`) built the greenfield
environment template and amended the incumbent template's header; Task 1
(`9a61145`, fix `f287fc0`) added the two shadow units and their
`UNIT_FILES` manifest row; Task 3 (`e22c252`, fix `c83c197`) added the
pre-create step and the timer-list comment; Task 4 (`6f698b1`) built the
render pass and the per-template install step; Task 5 (`25d3e4e`, fix
`47e80cc`) added the two build stamps, the `Makefile` plumbing, the
two-row version verb, and the smoke assertion; Task 6 (`4fa812f`) retired
five of the six ratified provisional marks. Every task was accepted by
`diff-reviewer`.

The post-chain round folded nine reviewer minors and retired the
launch-date wording from four shadow comment sites under Geoff's
no-calendar-schedule ruling (`a86475c`). Task 7's fold (`992e39c`) split
the smoke step's single assertion into two: an unrunnable binary now
fails with its own diagnosis, separate from the matcher-commit check, and
a new `scripts/check.sh` leg holds the three copies of the state
directory (the template, the shadow unit's `ReadWritePaths`, and
`deploy.sh`'s own variable) in agreement. A follow-up commit (`89920ba`)
pinned the version verb's test to a literal two-row expectation. The
`code-simplifier` agent reviewed Task 5's Go and changed nothing.

### Task 2's provenance line

`env/dubplate-server.env.template:55-60` forward-references this entry
for `DUBPLATE_SITE_NAME`'s copy provenance: the literal was copied
verbatim from the age store's `DUBPLATE_SERVE_SITE_NAME` by exit-status
comparison, and the value itself was never printed in any report or
commit message.

### The conductor's `--check` record

`bash scripts/deploy.sh --check 5.78.148.42` ran read-only on
2026-09-08 00:35 and exited 0 across sixteen announced steps.
`/etc/dubplate/env` reports no changes across its 42 variables.
`/etc/dubplate/server.env` reports would change across its 10 variables,
expected because the file does not exist on the box yet. The units rsync
lists exactly the two new units. The binary rsync lists one file. The
repo-config rsync with `--delete` would remove the orphaned
`/opt/dubplate/config/filebrowser.yaml`, dropped from the repo at
`b4295f1`.

### Task 7's verdict (fresh-context unit read, `claude-opus-5`)

The sitting may apply the units as they stand. The
`ReadWritePaths=/var/lib/dubplate/shadow` grant is the minimum, argued
statically from the store's own footprint. The hardening set matches
`dubplate-serve.service`'s directive-for-directive, with the `AF_NETLINK`
exclusion the one reasoned deviation, for a unit that runs no subprocess.
Both units verify under `systemd-analyze verify`. Three items fold at
close and carry into the sitting checklist below: the apply re-runs seven
pre-existing idempotent steps beyond this pass's five new artifacts; the
operator rule for a hand-run shadow verb; and the `DUBPLATE_SITE_NAME`
universality debt.

### The ten retired resolutions

Each provisional mark is now ratified. The location named is where the
application sits at HEAD.

- **R2** (five provisional marks retire): `internal/engine/shadow/rescore.go:28`
  and `match.go:20` carry ruling 3; `internal/engine/shadow/run.go:388`
  carries ruling 2; `cmd/dubplate-server/shadow.go:22` and `:37` carry
  rulings 4 and 5. `internal/engine/shadow/flac.go:62` stays as written,
  a rung 8 handoff mark rather than a pending decision.
- **R3** (the version verb prints two labeled rows): `cmd/dubplate-server/main.go:73-90`,
  `newVersionCmd`, printing `version` then `matcher-commit` in
  `doctor.go`'s own row shape.
- **R4** (the smoke assertion runs inside the single remote command):
  `scripts/deploy.sh:364-366`, one `run_remote` string that both a
  `--check` print and a real apply run verbatim.
- **R6** (`/etc/dubplate/server.env` is 0600 root:root): `scripts/deploy.sh:403-410`,
  the same mode and ownership as `/etc/dubplate/env`, with the reason
  stated in the comment at `:410`.
- **R7** (the masthead literal is copied, never invented):
  `env/dubplate-server.env.template:55-61`.
- **R8** (the floor and budget ship as literals): `env/dubplate-server.env.template:102-109`,
  `DUBPLATE_SHADOW_FLOOR=720h` and `DUBPLATE_SHADOW_BUDGET=60`, matching
  the registered defaults.
- **R10** (`TIMER_UNITS` keeps its list and gains a sentence):
  `scripts/deploy.sh:521-524`, the comment naming why
  `dubplate-shadow.timer` is absent from the unattended enable list.
- **R11** (the conductor runs `--check`, not the implementer): applied
  above, this entry's `--check` record.
- **R12** (doctor prints neither stamp): declined, so no file changes;
  filed as a note for whichever rung next touches `cmd/dubplate-server/doctor.go`.
- **R14** (the spec's stale citations are corrected at the close): the
  eight corrections are queued for the conductor's next spec-touching
  batch; `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
  is not edited by this pass.

### The sitting checklist (Task 9, Geoff's attended session)

Run in order, every `deploy.sh` invocation carrying its address so a
bare invocation never resolves the box through `hcloud`:

1. Confirm the box is quiet: no live executor, no warm uncommitted
   change the sitting did not author.
2. `bash scripts/deploy.sh --check 5.78.148.42` and compare its output
   against this entry's `--check` record above; a drift is a stop.
3. `bash scripts/deploy.sh 5.78.148.42`, the apply. It re-runs seven
   pre-existing idempotent steps beyond this pass's five artifacts (the
   `filebrowser.yaml` deletion, `systemd-tmpfiles --create`, four
   `mkdir`/`chown` steps, `docker compose up -d`, the four guardian
   timers' `enable --now`, and the unconditional
   `uv tool install --force beets==2.13.1` reinstall) as ordinary parts
   of the same script, not surprises.
4. `docker compose ps`, then one Navidrome reachability check, before
   any timer enable.
5. `systemctl enable --now dubplate-shadow.timer`: the act that starts
   the shadow's evidence clock.
6. Read the journal for the shadow's first pass and record whether it
   observed anything.
7. `ls -Zd /var/lib/dubplate /var/lib/dubplate/shadow` and
   `ls -Zd /var/lib/musicbox`, both SELinux verifications.
8. `df -h /`, the free-space baseline.
9. Read `started-at` off the box rather than off the sitting's wall
   clock; it is written on the first pass that completes all three
   steps without a fatal error, including a pass that observes nothing.
10. **Operator rule**: a hand-run shadow verb goes through
    `systemctl start dubplate-shadow.service` or `systemd-run` with the
    same unit properties, never the bare binary run as root.
11. Two non-blocking options remain open at the sitting's discretion:
    whether to write the rewrite post-mortem now or later, and whether
    to review the two albums waiting in
    `/srv/music/review/2026-08-31/geoff/` in the same session.

The floor the launch gate reads is an evidence minimum, not a calendar
date; this checklist and entry name no launch date and no calendar
arithmetic.

### Carried items

- The `DUBPLATE_SITE_NAME` universality debt: a site-specific literal in
  `env/dubplate-server.env.template`, retired when the name is re-minted
  through `secret-set.sh` and the row flipped back to bare.
- Doctor rows for both build stamps, filed for whichever rung next
  touches `cmd/dubplate-server/doctor.go` (R12).
- The retention knob's trigger: the report's store-size row against the
  sitting's free-space baseline (forward carry 8).
- Rung 9's re-verification of the limiter's caller-priority policy
  against a live measurement, which arrives only after the deploy
  (forward carry 6).
- The declared coverage-gap total gains four at this pass (the grant's
  minimality has no mechanical check; the environment template's values
  have no checker; the smoke assertion has never run against a box; the
  unit has never started), bringing the running total to thirty-four
  (rung 3's twenty-one, rung 4a's nine, rung 4b's four). Rung 15 reads
  the total rather than re-counting it.

### Spend

4b's unattended half ran about 2.6M against its 3M ceiling: the chain,
1.03M across 20 agents; Task 0, 0.09M; Task 7, 0.12M; the leftovers
round, the test pin, and the simplifier pass together, about 0.32M; plan
preparation, 1.06M; conductor turns, about 0.1M. The sitting added about
0.05M conductor turns. Attended interactions for the whole pass:
planning misses 0, execution sittings 1 (the sitting below, a planned
one, not a miss).

### The sitting, and the close

Task 9 ran 2026-09-08 with Geoff present and the conductor at the
keyboard, against the plan's Task 9 and "The sitting checklist" above.

**The quiet-box check**, 17:24 UTC: load 0.00, no running import, no
failed units, one entry in `/srv/music/inbox`, four guardian timers
scheduled.

**`deploy.sh --check 5.78.148.42`** matched the recorded 2026-09-08
00:35 output except transfer statistics, the rebuilt binary's size, and
the smoke assertion's improved shape from the Task 7 fold; no drift.

**The apply**, `deploy.sh 5.78.148.42`, exited 0 at about 17:25 UTC:
sixteen steps. The orphaned `/opt/dubplate/config/filebrowser.yaml` was
deleted by `rsync --delete`; the two shadow units and the binary were
installed; the smoke test printed `version pre-greenfield-321-gbe27434`
and `matcher-commit be274346add7abc339bd6ca0cf4c3b3071dd1fdf`;
`/etc/dubplate/env` was unchanged (42 variables); `/etc/dubplate/server.env`
was written (10 variables, mode 0600 root:root, 366 bytes);
`/var/lib/dubplate` and `/var/lib/dubplate/shadow` were created and
chowned `music:music`; `docker compose up -d` ran; the four guardian
timers were re-enabled; beets 2.13.1 was reinstalled via `uv` (with
uv's PATH warning, harmless).

**After the apply**, `docker compose ps` showed `musicbox-navidrome-1`
and `musicbox-cloudflared-1` both Up 6 days; Navidrome's `/ping`
answered HTTP 200 in about a millisecond; SELinux Enforcing with
`unconfined_u:object_r:var_lib_t:s0` on `/var/lib/dubplate`,
`/var/lib/dubplate/shadow` and `/var/lib/musicbox`, closing the
`/var/lib/musicbox` check carried open since the rewrite pass. `df -h /`
showed 75G size, 4.5G used, 68G available, 7%, the free-space baseline
for the response-store retention trigger.

**The shadow's evidence clock started at the timer enable.**
`systemctl enable --now dubplate-shadow.timer` ran at 17:25:56 UTC, its
first fire scheduled 17:30:26 UTC. The first pass ran at 17:30:30 UTC:
`Result=success`, exit 0; the journal line read `shadow observe:
captured 0, observed 0 (0 agreement(s), 0 divergence(s))`; the store
root was created with `observations/`, `passes/`, `pending/`,
`resolved/` (0750), `consumed-reports.json` (4770 bytes) and
`started-at`; no failed units. **`started-at` reads
`2026-09-08T17:30:30.116835324Z`**, written on the box by the first pass
that completes all three steps without a fatal error, and is the deploy
date as a first-class figure. The pass file is
`pass-20260908T173030Z.json`.

**The deploy date is 2026-09-08**, the apply at about 17:25 UTC and the
clock start at the `started-at` value above.

**Two discretionary items were left open at Geoff's discretion**: the
rewrite post-mortem, and the two albums in
`/srv/music/review/2026-08-31/geoff/`.

**Rung 4 is now closed end to end**, 4a and 4b both: the shadow contract
lands, the units run on the box, and the evidence clock is running
toward the launch gate's floor.

## Rung 4, pass 4a-ii (pass close, 2026-09-07)

Tasks 10 through 15 executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-05-rung-4a-shadow-runner.md` (revision
2.1), closing the fifteen-task 4a plan that pass 4a-i split at the Task
9 / Task 10 boundary. Rung 4a is now closed. The chain was
`wf_c97fd506-dda`: 16 agents, sonnet implementers and opus
`diff-reviewer`s, strictly sequential, 1.32M tokens.

### What landed

Task 10 (`09353f0`, fix `8326499`) added
`analysis.AdjudicationEntry.MatcherCommit` and `MatcherConfigHash`,
both `omitempty`, with the two `checkAdjudicationEntry` rules that
require the commit on a shadow-subject row and refuse either field on a
case-subject row, plus
`TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost`. Task 11
(`bef628e`) built `caseJSON`, `pendingJSON` and `storeArtifacts` in
`cmd/dubplate-server/shadow.go`, the marshalling that keeps
`internal/engine/shadow` clear of both harness shapes: `pendingJSON`
runs `analysis.CheckAdjudications` before it returns bytes, so a
malformed row never reaches disk. Task 12 (`81183fe`, fix `3c229f0`)
built `shadow.BuildReport` and the `shadow` subcommand's three verbs,
`observe`, `report` and `reevaluate`, in `newCheckCmd`'s shape, with
`main.matcherCommit` defaulting to `"dev"` and threaded through
`newShadowCmd`. Task 13 (`f7cd7eb`, fixes `f7f82b7` and `0fed5b3`)
built `shadow.Reevaluate`, the mechanical re-evaluation of an
observation stamped at a superseded matcher commit: it replays that
observation's own stored responses through `replayTransport` and
`matchAlbum` under `musicbrainz.Unmetered()` at the observation's own
`EvalDate`, and sorts every observation into `Retired`, `Restamped`,
`NewlyDiverged`, `Unevaluable` or `Skipped`. Task 14 (`9c956b8`, fix
`bb75cea`) built rung 4's exit gate, one end-to-end `observe` run over
a hand-authored fourteen-directory inbox. The post-chain round
(`85dcd37`) folded the Task 13 and Task 14 review leftovers.

The close-out `code-simplifier` pass (`2992bbb`, 9 files, +140/-217)
exported `shadow.StoreDir` to retire a twice-written directory literal,
folded four identical merge loops in `BuildReport` onto one helper,
collapsed `Reevaluate`'s three outcome arms onto one shared
restamp-and-write block, replaced a hand-rolled map comparison with
`maps.Equal`, and replaced collect-then-sort and membership loops with
`slices.Sorted(maps.Keys)` and `slices.Contains`. Simplify review:
accept, no blockers; the reviewer confirmed the outcome switch keeps the
write order and the agreement-to-agreement skip, that `maps.Equal` is
strictly stricter than the helper it replaced, and that the restamp
assertions now read from disk rather than the mutated pointer.

Also landed this session, outside the pass but inside this ledger's
window: the streamer 2.0 initiative. Its design spec
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision
2.1) is ratified after four reviews, `ROADMAP.md` carries the
`streamer-2-0` initiative, and CLAUDE.md carries the build-for-2.0 and
prior-art rules, the build spec's 2026-09-07 erratum, and the SvelteKit
static web rule. The repo also gained an MIT `LICENSE`.

### What the gate caught

Two tasks accepted on the first review, Task 11 and Task 12's second
round. Task 10 drew a `fix`: the `ReleaseMBID` doc comment landed only
half of owner ruling 2, naming the field as the release under dispute
and dropping "filled by whichever side proposed one," which is the
clause that answers which MBID goes in the field. Task 12 drew a `fix`
for a printed row rather than for logic: the report emitted the
`go-applies-oracle-skipped` row with its off-box note but omitted the
divergences-by-disposition row the criteria name, and the absence was
explained only in a Go doc comment an operator never sees.

Two tasks went past the one-re-dispatch rule and were ruled by the
conductor.

- **Task 13** drew a `fix` for a comment misattachment, landed as
  `f7f82b7`, then a second `fix` on a real assertion gap: nothing read
  `resolved/<obs-id>.json` back, and the reviewer proved the gap by
  swapping `OriginalCommit` with `RetiringCommit`, `OriginalClass` with
  `NewClass` and `OriginalGo` with `NewGo` all at once with the package
  still green. The conductor ruled the second fix rather than accepting
  a deliverable no test could fail on. `0fed5b3` reads the resolved
  record back, restamps a retired observation so the per-commit report
  row drains, merges the `PendingRowFor` pass-through into one exported
  function, and dates the resolved record with `time.DateOnly` to match
  the pending rows. Its review accepted, with three swap mutations run,
  each firing two assertion misses.
- **Task 14** **escalated**. Named mutations 4 and 6 did not fire, and
  the implementer had written the mutation proofs as design commentary
  at each defending assertion rather than running them. The reviewer
  ran four of the six itself and reported the two that were unfired.
  The conductor ruled a fix on `opus`, landed as `bb75cea`: all six
  mutations run, firing 7, 8, 4, 2, 3 and 2 assertion misses in order.
  The wall-time bound is a `time.AfterFunc` timer whose `Stop` return
  is the assertion, because `TestEngineNeverReadsTheClock` bars
  `time.Now` from every test under `internal/engine`. Determinism now
  builds two trees rather than one. Row 04 is un-tag-logged, so its
  review-tree collision suffix is load-bearing and the test asserts
  which reader produced its oracle signal.

**Conductor rulings taken this pass.**

1. **B7 stands.** `newShadowCmd` keeps the injected `matcherCommit`
   parameter. Resolution R3 rejects the package variable the
   alternative would have to write.
2. **Mutation proofs are run, never narrated.** A named mutation is
   applied to the source, the suite is run red, and the source is
   restored. This is now written into
   `.claude/agents/dubplate-implementer.md`, the repo's own executor,
   so it binds every later rung rather than this pass alone.
3. **The exit gate's figures are 14 directories, 3 uploaders and 13
   stored observations.** The plan's single-uploader fixture is
   unachievable alongside its own row-04 backstop
   (`reconcile.go:319-323`) and its per-uploader partial-skip flag
   (`reconcile.go:103-106`, `:325-328`). Uploader `pk` carries row 09
   and the ambiguous-basename partner, `bo` carries row 07, `ada`
   carries the other eleven.
4. **Retired observations are restamped; agreement-to-agreement
   observations are not.** The asymmetry is recorded rather than
   changed: a retire has a report row to drain, and a replay that
   agrees the way it already agreed has nothing to record.
5. **The plan's mutation-6 narration is a half-truth.** Dropping
   `buildInbox`'s `os.Chtimes` calls fails
   `TestRunIsDeterministicUnderAFrozenClock` only.
   `TestObservePassIsIdempotent` is insensitive to mtime pinning,
   because a second `Run` inside one process reads the same unpinned
   mtimes.

### The shadow contract coverage, S1 through S23

The plan's contracts, each with the Go site that implements it and the
test that proves it. A contract no end-to-end run can reach names the
hand-authored fixture instead.

| # | Contract | Go site | What proves it |
| --- | --- | --- | --- |
| S1 | `tag.Album` is the fixed entry point | `shadow/match.go` `matchAlbum` | `TestMatchAlbumTakesTheIDConsensusEarlyReturn`, `TestMatchAlbumSearchesWhenThereIsNoConsensus` |
| S2 | The estate tunables are `QuietImportTunables()` | `matcher/tunables.go`, `shadow/match.go` | `TestMatchAlbumRunsUnderFullCompatAndQuietImportTunables`; `conform`'s drift guard against `gradeTunables("estate")` |
| S3 | The oracle's signal is binary | `shadow/classify.go` | `TestClassifyVerdictReadsTheStoredAction` |
| S4 | Four outcomes, three of them divergences | `shadow/classify.go` | `TestClassifyEveryOutcome`; all three classes in `TestObservePassEndToEnd` |
| S5 | Four deferring branches, closed deliberately | `shadow/classify.go` | `TestClassifyUsesErrorsIsRatherThanEquality`; `TestObservePassEndToEnd` (unavailable, budget), `TestObservePassDefersEverythingOnAClosedLimiter` (limiter closed) |
| S6 | No error but `ErrNoCandidates` is a Go skip | `shadow/classify.go` | `TestClassifyEveryOutcome`'s deferral rows |
| S7 | The capture digest is the identity | `shadow/capture.go` | `TestCaptureDigestMovesWithEveryFileProperty`, `TestCaptureIsIdempotentUnderAnUnmovedDigest`, `TestHasObservationIsTheIdempotenceCheck` |
| S8 | A partial read is discarded, never stored | `shadow/capture.go` | `TestCaptureDiscardsAPartialRead`, `TestCaptureTreatsAnUnreadableAlbumDirectoryAsAPartialRead` |
| S9 | The grouping rule, with beets' artist fallback | `shadow/capture.go` | `TestCaptureExcludesAMultiAlbumDirectory`, `TestCaptureExcludesAFlatArtistOnlyMultiAlbumDirectory` |
| S10 | The join key is the basename, never `SourceDir` | `shadow/reconcile.go` | `TestReconcileJoinsAnImportedRowByContentNotBySourceDir`, `TestReconcileResolvesASkipFromEitherBackstop`, `TestReconcileExcludesGateOutcomes`, `TestReconcileClaimsGateRejectedOnlyThroughReviewTree` |
| S11 | Four `unreconciled` reasons | `shadow/reconcile.go` | `TestReconcileUnreconciledReasons`, `TestReconcilePartialSkipIsUnreconciledNotASkip`; all four counted in `TestObservePassEndToEnd` |
| S12 | The reports are a rolling window, and gaps are named | `shadow/oraclereport.go`, `shadow/reconcile.go` | `TestReadRunReportsSortsByStartedAt`, `TestReconcileDetectsAReportGap`, `TestReconcileNoGapWhenReportsStillPresent` |
| S13 | The store layout, rooted at one directory | `shadow/store.go` | `TestStoreRoundTripsEveryArtifact`, `TestStoreWritesNothingOutsideItsRoot`, `TestStoreLeavesNoTempFile` |
| S14 | `passes/` exists because a `Budget` is in-process state | `shadow/store.go`, `shadow/report.go` | `TestBuildReportSumsTheStore` (11, not 3 and not the maximum), `TestPassIDCollisionSecondWriteWins` |
| S15 | The budget, the pace and the contact | `shadow/config.go`, `cmd/dubplate-server/shadow.go` | `TestDefaultIntervalIsTwiceMusicbrainzDefault`; the cap asserted in `TestObservePassEndToEnd` against `observeBudgetCap` |
| S16 | Mode one records, mode two re-scores | `shadow/rescore.go` | `TestRescoreDefaultScoresUnderDefaultCompat`, `TestRescoreDefaultMarksItsVerdictAsAHypothesis` |
| S17 | Any replay client takes `Unmetered()` | `shadow/rescore.go`, `shadow/reevaluate.go` | `TestRescoreDefaultIssuesNoSearchAndSpendsNoToken`, `TestRunSpendsNoTokenOnTheRescore`, `TestReevaluateSpendsNoBudgetAndReachesNoNetwork` |
| S18 | The replay transport refuses a miss with 501 | `shadow/transport.go` | `TestReplayTransportRefusesAMissWithoutA404`, `TestRescoreDefaultReportsAMissRatherThanDroppingACandidate`, `TestReevaluateReportsUnevaluableRatherThanRetiring` |
| S19 | The engine marshals neither harness shape | `cmd/dubplate-server/shadow.go` | `TestCaseJSONRoundTripsThroughLoadCase`, `TestStoreArtifactsWritesOneFilePerObservationAndOnePerNewsRow`; `conform`'s `TestEngineNeverImportsHarness` |
| S20 | Every staged row is validated before it is written | `cmd/dubplate-server/shadow.go` `pendingJSON` | `TestPendingJSONValidatesBeforeItReturns` (four-row table), `TestPendingJSONNeverEmitsATodo` |
| S21 | The two schema additions and their rules | `analysis/adjudication.go` | Three inline table rows (`shadow-subject-no-commit`, `case-row-with-matcher-commit`, `case-row-with-matcher-config-hash-only`) plus `TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost` |
| S22 | Two clauses gate, not one | `shadow/report.go`, `cmd/dubplate-server/shadow.go` | `TestBuildReportSumsTheStore`, `TestReportPrintsThePinnedRowForm`; the second clause's disposition join is declared off-box (gap 4) |
| S23 | `disposition` stays owner-only | `cmd/dubplate-server/shadow.go` `pendingJSON` | `TestPendingJSONValidatesBeforeItReturns`'s `OracleWrong` row; the shadow writes into the committed ledger never |

### The declared coverage gaps

Nine, on one criterion: a surface 4a itself builds that no test in this
pass can reach, or a real input shape the design deliberately does not
observe. None blocks.

1. **An archive upload is invisible to the capture** (Resolution R19).
   `stageArchive` extracts a `.zip` and removes it from the inbox, so
   the album that later appears in a run report has a basename the
   shadow never saw.
2. **The re-scored `DefaultCompat` verdict is a hypothesis, not a
   measurement**, wherever the real flow would have widened the
   candidate set. That is ruling 3's own limit, and the row says so.
3. **`Budget.wait` cannot be injected per album** (Resolution R7), so
   the deferral branches are proven in two runs rather than one.
4. **The report's disposition rows cannot be computed on the box**
   (Resolution R21). Owner dispositions live in the corpus on the
   workstation, and the launch gate does that join.
5. **`atomicfile`'s temp-directory property is unproven on a
   single-filesystem workstation**, held by the doc comment plus 4b's
   `ReadWritePaths` argument.
6. **No test exercises a real library tree.** `readFLAC` runs against
   an injected reader in the reconcile tests and against built FLACs in
   the end-to-end run. A real beets-written library file is first read
   at 4b's deploy.
7. **The shadow's own timing against a live import is untested.** The
   quiescence rule is a constraint on 4b's timer interval, and 4a
   proves no part of it.
8. **"`internal/engine/shadow` imports `internal/config` never" has no
   mechanical guard.** Conductor ruling C3 keeps this a declared gap
   with the `diff-reviewer` as the meanwhile guard, rather than
   widening this pass's `conform` territory.
9. **The cross-filesystem half of `atomicfile`'s rename property is
   untested**, distinct from gap 5's temp-directory half. Nothing
   proves what `os.Rename` does across an EXDEV boundary.

Joined to rung 3's twenty-one, **the running total is thirty**. Rung 15
reads that number rather than re-counting.

### Evidence

`bash scripts/check.sh` passed clean at every commit's review point.

The exit gate, `TestObservePassEndToEnd`, drives one `observe` pass over
fourteen album directories under three uploaders and asserts parsed
structure only, with no golden file anywhere:

- 2 agreements (row 01's ID consensus, row 02's empty search).
- 1 divergence of each class: `go-skips-oracle-applied` (03),
  `go-applies-oracle-skipped` (04), `go-applies-different-release`
  (05).
- 2 excluded: 1 multi-album (06), 1 gate-rejected (07).
- 4 unreconciled reasons, 5 rows: no release id (08), partial skip
  (09), ambiguous basename (10 and its structural partner), report
  swept (11).
- 2 deferred: `ErrUnavailable` (12), `ErrBudgetExhausted` (13).
  `ErrLimiterClosed` is proven separately by
  `TestObservePassDefersEverythingOnAClosedLimiter`, per gap 3.
- 3 pending rows, one per divergence, each carrying a `release_mbid`.
- 13 stored observations: the fourteen captured directories less row
  06, which capture excludes before it ever becomes an observation.
- `started-at` holds the pass clock, and `ClockStarted` is true.

**The budget arithmetic.** `observeBudgetCap` is 12, a fixed literal
rather than a figure derived from the fixture table: rows 01 through 05
reserve one token each, and row 12's seven exhausted retry attempts
spend one each, for 5 + 7 = 12. Row 13's reserve is blocked and spends
nothing. `Budget.Spent` is asserted equal to the cap, so the cap is a
cap and not a decoration.

**Measured wall times.** `go test ./...` 2m15s to 2m27s;
`internal/harness/conform` 132.8s to 143.0s; `internal/engine/shadow`
0.175s to 0.2s; `bash scripts/check.sh` 1m6s to 3.5 minutes. The
`internal/engine/shadow` figure is two orders of magnitude inside the
five-second wall bound the exit gate asserts.

**The conform figures are untouched by this pass**, read from the 4a-i
entry rather than re-measured: tier 0 still replays 72/72, translation
still grades 50/50, the request tier still runs 215/303/48, and tier 2
still produces 215 verdicts. Nothing in `internal/engine/shadow`
reaches the conform tree.

### The mutation table

Every named mutation across Tasks 10 through 14, each applied to the
source, run red, and reverted.

| Task | Mutation | Outcome |
| --- | --- | --- |
| 10 | Drop the shadow-row `MatcherCommit` requirement | fires, `shadow-subject-no-commit` |
| 10 | Narrow the case-row guard to `MatcherCommit` only | fires, `case-row-with-matcher-config-hash-only` |
| 10 | Drop `omitempty` from `MatcherCommit` | fires, the ledger round trip |
| 11 | Skip the `CheckAdjudications` call | fires, the three refusal rows |
| 11 | Substitute `analysis.OracleWrong` for `Pending` | fires, two `checkAdjudicationEntry` rules |
| 11 | Emit a literal `TODO` for `ReleaseTruth` | fires, `TestPendingJSONNeverEmitsATodo` |
| 11 | Corrupt `PathQuery` to a non-canonical form | fires, `cassette.LoadCase` rejects |
| 12 | Let `observe` run on a `"dev"` stamp | fires, `TestObserveRefusesAnUnstampedBuild` |
| 12 | Report the last pass's budget instead of the sum | fires, `TestBuildReportSumsTheStore` |
| 12 | Call `time.Now` inside `BuildReport` | fires, `TestEngineNeverReadsTheClock` |
| 12 | Print map rows in map order | fires, `TestReportPrintsMapRowsInSortedKeyOrder` |
| 12 | Print to `os.Stdout` rather than the injected writer | fires, `TestShadowVerbsAreRegisteredAndPrintToStdout` |
| 13 | Answer a replay miss with 404 rather than 501 | fires, `TestReevaluateReportsUnevaluableRatherThanRetiring` |
| 13 | Drop the newly-diverged arm | fires, `TestReevaluateRecordsANewDivergenceOnAnObservationThatAgreed` |
| 13 | Re-run at `now` rather than the stored `EvalDate` | fires, `TestReevaluateRetiresADivergenceThatNoLongerReproduces` |
| 13 | Give the replay a metered budget | fires, `TestReevaluateSpendsNoBudgetAndReachesNoNetwork` |
| 13 | Restamp an unevaluable observation | fires, `TestReevaluateReportsUnevaluableRatherThanRetiring` |
| 13 (fix) | Swap the resolved record's commits, classes and verdict pairs, three swaps | each fires, 2 assertion misses each |
| 14 | Make `capture` skip the grouping check | fires, 7 misses |
| 14 | Make `reconcile` treat a `rejected` row as a skip | fires, 8 misses |
| 14 | Set the budget cap to zero | fires, 4 misses |
| 14 | Drop `Retry-After: 0` from the test server | fires, 2 misses, through the `AfterFunc` bound |
| 14 | Reverse the fixture's numeric prefixes | fires, 3 misses |
| 14 | Drop `buildInbox`'s `os.Chtimes` calls | fires, 2 misses, `TestRunIsDeterministicUnderAFrozenClock` only |

Four mutations named in 4a-i stay unfired, each with the guard that
holds the property instead, and none is re-litigated here: Task 1's
mutations 1 and 3, Task 5's mutation 2 and Task 8's mutation 2. The 4a-i
entry carries their reasons.

### Non-blocking leftovers, condensed

- `printCountRows` prints nothing for an empty map, so a launch-gate
  reader cannot tell "zero divergences" from "this report does not
  compute divergences." Every vocabulary is a closed constant set, so
  printing the known keys at zero is available.
- The store path literal `filepath.Join(cfg.StateDir, "shadow")` is
  written in `cmd` twice, duplicating the engine's own unexported
  `storeDirName`. An exported `shadow.StorePath(stateDir)` would keep
  the layout in one place.
- `report` calls `OpenStore`, which creates the store directory as a
  side effect of merely printing a report on a box that has never run
  `observe`.
- The printed `go-applies-oracle-skipped` count is the raw count, while
  the gated figure is the subset the owner has not disposed
  `oracle-wrong`. The parenthetical carries the caveat and the raw
  count can only over-report, so it fails safe.
- `Report.MatcherCommits` counts non-agreement observations only, so a
  commit observed under with zero divergences never gets a row.
- Any replay error in `Reevaluate` that is neither a recording miss nor
  `tag.ErrNoCandidates` aborts the whole call, discarding outcomes
  already computed. `classify` defers per observation instead.
- `stageNewlyDiverged` and the `reevaluate` verb's `RunE` have no test
  in `cmd/dubplate-server`; `shadow_test.go` asserts only that the verb
  is registered.
- The four committed `testdata/observe` bodies are byte-for-byte what
  the in-package `releaseBody` and `searchBody` helpers already
  generate, so they are a second source of truth that can drift.
- 4a-i's own leftover list stands unchanged: no directory fsync, the
  non-`.json` decode in `Observations()` and `Passes()`, the uncapped
  consumed ledger, the silent content-join drop, the unsynchronized
  `recordingTransport`, and the terminal deferral.

### Filed for other owners

- Renaming `case.json`'s `Profile` the day the corpus grows a
  `quiet-import` profile, for lane E1 (Resolution R8).
- A `KindFloat` threshold variant, for the first rung 3 follow-on after
  rung 1b lands (forward carry 3).
- The 4a-i filings stand: rung 8 absorbs the FLAC reader, rung 16
  retires the oracle report decode, and rung 1b absorbs
  `integrationClasses` as data.

### What a later pass would be wrong to rediscover

1. **Mutation proofs are run, never narrated.** A mutation described in
   a comment beside the assertion it is supposed to defend proves
   nothing, and Task 14 shows why: two of six such mutations did not
   fire, and no reader could tell from the commentary. The rule is now
   in `.claude/agents/dubplate-implementer.md`.
2. **The exit gate's uploader count is forced by the code, not by
   taste.** `reconcile.go` scopes the partial-skip flag per uploader
   and lets the first review-only capture under an uploader consume the
   gate-rejected count, and an ambiguous basename is by construction a
   collision between two uploaders. Any fixture that wants a clean
   partial skip, a clean gate rejection and an ambiguous basename in
   one pass needs three uploaders.
3. **`TestEngineNeverReadsTheClock` covers tests too.** Nothing under
   `internal/engine` may call `time.Now`, so a wall-time bound in an
   engine test is a `time.AfterFunc` timer whose `Stop` return is the
   assertion, never a stopwatch.
4. **Retire restamps, agreement-to-agreement does not.** A retired
   observation is restamped so `BuildReport`'s per-commit row drains
   and a second `reevaluate` does not re-retire it. An observation that
   agreed and still agrees is left untouched, because there is no row
   to drain. The asymmetry is deliberate.
5. **`os.ReadDir` mtime granularity is the digest's assumption.** The
   capture digest ingests modification times, so a fixture that does
   not pin mtimes with `os.Chtimes` produces a different digest on
   every run. The determinism test is what catches that, and it needs
   two separately built trees to catch it at all.
6. **A retire is only idempotent because the observation is
   restamped.** `RemovePending` tolerates `ENOENT` and `PutResolved`
   overwrites, so without the restamp a second pass rewrites the
   resolved record with a fresh date and reports the retire again.

### The rung 4 spec corrections, applied at this close

Each is marked in the spec as "(corrected 2026-09-07 at the 4a-ii
close)": `spec:1608-1609`'s "(ruling 7)" reads ruling 6, the deploy
sitting; `spec:299-301`'s pre-fold one-request-per-second sentence
follows the conductor's pace ruling; `spec:1267`'s "the four
`DUBPLATE_SHADOW_*` rows" reads six; `spec:845-848`'s byte-for-byte
claim reads as the round trip that actually holds, and its fixture form
reads as an inline Go literal variant; `spec:1616`'s exit gate names
only the deferral branches one pass can reach; and `spec:1063-1066`'s
`newShadowCmd` signature carries the matcher stamp. The `spec:441`
grouping-key correction was already applied at the plan's own fold and
needed nothing here. Rung 3d's own "rung 3 to rung 4 handoff"
section carries the matching wire-pace amendment.

### The six owner rulings, ratified

Geoff ratified all six on 2026-09-07 (`57b49fd`), each as recommended:
the integration-class wall, the release under dispute filled by
whichever side proposed it, both compat modes, the 720h floor, the
60-request cap at twice the limiter's own interval, and the deploy as an
owner sitting. **The pass's batched owner ask is therefore answered
rather than open.** What remains is mechanical: every provisional
application still carries a "provisional, owner ruling N open" marker in
the landed code, and retiring those markers is a search rather than a
re-reading. 3d's inherited ask, re-dispositioning
`bug-assignment-rejection-gap` from `oracle-wrong` to
`no-correct-answer`, rides its own stream and is not rung 4's to raise.

### The 4a to 4b handoff

What 4b's plan needs from 4a, stated rather than derived.

1. **The store's root and its footprint.** `<state dir>/shadow/` holds
   `passes/`, `observations/<obs-id>/`, `pending/`, `resolved/`,
   `started-at` and `consumed-reports.json`, and nothing is written
   anywhere else. `TestStoreWritesNothingOutsideItsRoot` proves it, so
   that root is 4b's `ReadWritePaths=` argument, static rather than
   assumed.
2. **The unit's `ExecStart` and its refusal.** `dubplate-server shadow
   observe` errors rather than observing when `main.matcherCommit` is
   empty or `"dev"`, and creates no store on either refusal. No build
   path sets the stamp today. **4b's deploy edit 6 is what makes the
   refusal observable**: without the two `-ldflags` stamps and the
   smoke-test assertion, an unstamped binary deploys clean, every timer
   pass refuses, and the `OnFailure=` chain fires into a slug whose URL
   is unset and therefore no-ops. `reevaluate` carries the same
   refusal.
3. **The six config rows and how they render.** The three paths are
   required-when-enabled and abort `Load` when the flag is set without
   them. The flag and the two thresholds are accepted and never
   required. An explicit `DUBPLATE_SHADOW_ENABLED=false` loads clean
   with a nil `Shadow`, which is the adopter's own way to turn the
   integration off. The greenfield template carries six
   `DUBPLATE_SHADOW_*` rows, not four.
4. **The wire pace and the cap.** `shadow.DefaultInterval` is
   `2 * musicbrainz.DefaultInterval` and the cap is sixty tokens, both
   now ratified under ruling 5 and both one literal away from a
   different value. Sixty tokens at two seconds a token bounds a pass at
   two minutes, comfortably inside 4b's five-minute timer.
5. **The operator rule the unit's comment must carry.** A hand-run
   shadow verb goes through `systemctl start dubplate-shadow.service`
   or `systemd-run` with the same properties, never the bare binary as
   root. The read-only guarantee is the unit's, not the binary's.
6. **The clock's definition** (Resolution R11). `started-at` is written
   on the first pass that completes all three steps without a fatal
   error, including a pass that observes nothing, so the clock starts at
   the deploy rather than at the first upload. **4b's close records that
   date as a first-class figure**, the fact the launch gate's floor
   check reads.

### Forward carries

1. **The rung 15 re-run criterion** travels forward unchanged, as it
   has from every rung 3 pass close, now joined by the shadow's own
   launch-bar clause.
2. **The declared coverage gaps are a running total of thirty**, rung
   3's twenty-one plus 4a's nine. Read it, never re-count it.
3. **The four-knob wall registration** stays the named follow-on for
   the first pass after rung 1b, and **rung 4's own `KindFloat` need
   joins it**. `Threshold.Kind` now has two values and the three float
   rec thresholds want a third. `AdminEmail`'s dedicated contact row
   sits beside both.
4. **Rung 1b absorbs `integrationClasses` as data**, not as a shape to
   invent, and the table's `Required`/`Optional` split is the semantics
   it must preserve.
5. **Rung 8 absorbs the FLAC reader** into `internal/engine/tagio` and
   repoints the shadow. **Corrected, pass 8b's close, 2026-09-08**: the
   mapping has eighteen fields, not nineteen (`DataSource` is never a
   file tag), and it lives in the 4a plan (`:2453-2478`) and inline at
   `flac.go:210-228`, not in the 4a-i entry; rung 8 does not re-derive
   it. The deployed binary carries landed-but-unwired packages from
   other lanes, and none is reachable from the `shadow` subcommand's own
   call graph; repointing is rung 8's task, never a silent adoption.
6. **Rung 9 re-verifies the limiter's caller-priority policy against a
   measurement.** 4a is the first `ClassBackground` caller and 4b puts
   it beside a live import. 4a's own figure is the exit gate's spend of
   twelve tokens; the live one arrives after 4b's deploy.
7. **Rung 6's store supersedes none of this**, and `internal/lock`
   stays unadopted. The timer path is a systemd singleton, `reevaluate`
   writes only under `resolved/`, `report` opens nothing for writing,
   and `observe` only ever creates.
8. **The response store is a consequence of ratified ruling B rather
   than a preference.** The retention knob is named rather than built:
   the trigger is the report's own store-size row against the root
   filesystem's free space, which 4b's sitting records as a baseline.
9. **The streamer 2.0 forward hooks are binding on rungs 5 through 9
   and rung 12**, through CLAUDE.md's build-for-2.0 and prior-art rules
   rather than through this pass's own carries.

**Resume prompt for the next session.** Rung 4b is next, in the main
tree: the deploy, whose close needs an owner sitting per ruling 6. Its
plan is owed just-in-time and does not exist yet.

**Budget.** 4a-ii spent about 2.7M against its 3M ceiling (Geoff,
2026-09-07), an estimate from the agent meters: staleness pre-task and
chain-args extraction 0.26M, chain 1.32M, the Task 13 and Task 14 fix
and review rounds 0.49M, simplify 0.18M plus its review, the close
drafter 0.16M, and the conductor's own turns for the pass about 0.2M.
The 2.4M flag was crossed at the close, after every task had landed,
by the simplify and close rounds; nothing was cut. The 2.0 brainstorm's
reviews and fold (about 0.9M) ran in the same session and are not
counted here. The chain alone was 1.32M across 16 agents. Attended
interactions for 4a-ii: planning misses 2 (the exit gate's
single-uploader fixture figure, and the mutation-6 narration, both of
which a probe at planning would have caught), execution sittings 0. The
conductor asked Geoff nothing during execution.

## Rung 4, pass 4a-i (intermediate close, 2026-09-05)

Nine plan tasks executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-05-rung-4a-shadow-runner.md` (revision
2.1), which splits the fifteen-task 4a plan at the Task 9 / Task 10
boundary on CLAUDE.md's pass-sizing rule. This entry is that
intermediate close, not the pass's own close: Tasks 10 through 15 land
in pass 4a-ii, a later session.

### What landed

Task 1 (`bc21683`, fix `a1d3d6c`) re-landed `internal/atomicfile`
unchanged from `_salvage` on merit, plus the plan's crash-safety test
and a follow-up subtest pinning the temp file's directory. Task 2
(`46f4229`, fix `751bdde`) built `matcher.QuietImportTunables`, the
engine-side constructor for the four quiet-import assignments, with
`conform`'s drift guard against `gradeTunables("estate")`. Task 3
(`f20f329`, fixes `afeb007` and `8b73af8`) built the integration-class
wall (`IntegrationClass`, `Config.Shadow`, required-when-enabled
resolution), `ThresholdKind` and `ParseDuration`, and doctor's
integration and duration rows. Task 4 (`adcf66d`) built
`internal/engine/shadow`, with the provisional FLAC container reader
and the nineteen-field mapping onto `matcher.ItemFields`. Task 5
(`5966b8a`, fix `60221e9`) built the shadow store, one directory per
observation under `observations/`, `passes/`, `pending/`,
`resolved/`, every write through `atomicfile.WriteFile`. Task 6
(`d1603ec`, fix `891dc1a`) built capture, the lexical
uploader/basename walk, the grouping rule, the digest, and the
partial-read discard. Task 7 (`b0aef65`, fixes `d1e09dc` and
`6bbcbe0`) built the oracle report decode and the reconcile step,
joining captures to run reports by content and basename rather than
by `SourceDir`. Task 8 (`4c8bcc4`, fix `ffabf71`) built the recording
and replay transports, `shadow.Config`, and `matchAlbum`, plus the
wire-pace constant `DefaultInterval`. Task 9 (`817c75f`) built the
`DefaultCompat` re-score, the divergence classification, and `Run`'s
five-step pass loop. The close-out `code-simplifier` pass (`ea1d0ff`,
3 files, +12/-34) folds three collect-then-sort loops onto
`slices.Sorted`/`SortedFunc`, replaces a hand-rolled newest-capture
scan with `slices.MaxFunc`, and drops a test helper that reimplemented
`strings.Join`. Simplify review: accept, no blockers; every rewrite
verified against the go1.27.1 stdlib and the newest-capture tie rule
confirmed unchanged.

### What the gate caught

Five tasks needed a fix round from the in-chain `diff-reviewer`, two
of them twice. Task 2 attached a citation to the wrong clause. Task 3
carried a stale citation, then an unproven budget threshold. Task 5's
test comment claimed coverage the store's own naming scheme cannot
provide. Task 6 aborted the whole pass on an unreadable album
directory instead of recording a partial read. Task 7 carried a stale
provisional-ruling comment, then a gate-rejected capture claimed by
the wrong row.

Two tasks escalated to the conductor rather than taking a
`diff-reviewer` fix.

- **Task 1** flagged that the plan's own mutation-1 narrative was
  wrong: moving `os.CreateTemp` to `os.TempDir()` is not caught by
  either the leftover-temp-file test or the new crash-safety test,
  because `os.Rename` needs write permission on the destination
  directory regardless of where the temp file lives. Ruled: add one
  extra test beyond the plan's fold, pointing `TMPDIR` at an
  unwritable directory. Landed as `a1d3d6c`.
- **Task 8** escalated on two points: whether `shadowScoring` is the
  legitimate mutation site for named mutation 6, and whether the wire
  pace and the budget cap owed a materialized constant rather than a
  doc comment. Ruled: `shadowScoring` is the accepted mutation site;
  `DefaultInterval = 2 * musicbrainz.DefaultInterval` is a named
  constant with a pinning test, landed as `ffabf71`; the 60-token cap
  stays Task 3's registered threshold, and wiring it into the
  `ClassBackground` budget is Task 12's job in 4a-ii.

The reviewer found one race and one misattributed claim beyond the
review's own blocking-list shape. Task 6's reviewer traced a genuine
race with `beets`: `review.Move` renames the album directory itself,
so a walk that lists a basename can take `ENOENT` reading that
directory's own entries, and the first cut aborted the whole capture
pass on that error instead of recording a partial read. Task 7's
second review found the gate-rejected exclusion claimed by whichever
capture under an uploader matched first, so a beets-skipped capture
could steal the exclusion meant for a genuinely rejected one and the
rejected album would surface a fabricated oracle skip signal; fixed
in the post-chain round as `6bbcbe0`.

Three mutations were recorded as unprovable rather than papered over.
Task 1's mutation 1 (moving the temp file out of the target's own
directory) was unfired by the original test pair, as above, before
the extra subtest closed it. Task 5's mutation 4 (dropping the
explicit sort in `Observations`) cannot fail with this store's naming
scheme: `os.ReadDir` already returns filename order and the directory
name is the id, so no test tree can make the two orders diverge.
Task 9's mutation 5 (handing the replay client the pass budget) is
unreachable without a signature change the plan's own fold predicted;
proven only by a temporary signature change that confirmed the effect,
then reverted.

### Evidence

`bash scripts/check.sh` passed clean at every commit's review point.
`go test ./...` runs about 2m06s, dominated by `internal/harness/conform`'s
sweep (unchanged from rung 3). The conform sweep's own figures (72/72,
50/50, 215/303/48, 215/294/294/214, 215 verdicts) are untouched by this
pass; nothing in `internal/engine/shadow` reaches the conform tree.

### The six owner rulings, applied provisionally

Each is written into the landed code as "provisional, owner ruling N
open." One line each, with what reverses if Geoff rules the other way.

1. **Ruling 1** (integration-class shape): option (a), the
   `IntegrationClass` table, is built. Reversing to (b) or (c) retires
   `internal/config`'s integration-class wall and moves the shadow's
   paths onto flat accepted keys or bare cobra flags.
2. **Ruling 2** (a skipped album's `ReleaseMBID`): filled by whichever
   side proposed a release, oracle's or Go's. Reversing to option (b)
   relaxes `checkAdjudicationEntry`'s pair rule for shadow subjects
   specifically.
3. **Ruling 3** (which `Compat` the shadow runs): both, `FullCompat`
   scored and `DefaultCompat` re-scored from the same candidate set.
   Reversing to (a) or (b) drops the re-score step and Task 9's
   classification collapses to one verdict pair.
4. **Ruling 4** (the estate's shadow floor): the shipped 720h default.
   Reversing changes only the registered threshold's default; no code
   branches on the value.
5. **Ruling 5** (the MusicBrainz budget cap): 60 requests per run at
   `2 * musicbrainz.DefaultInterval`. Reversing to `DefaultInterval`
   moves the shadow's wire pace back to one request per second and
   reopens the shared-quota mitigation question ruling 5 itself
   resolves.
6. **Ruling 6** (the first deploy, sitting or conductor action): an
   owner sitting. This ruling touches only 4b, out of this pass's
   scope; nothing in 4a-i depends on its answer.

### Conductor rulings taken

C1 through C3 landed at the plan's fold, before any task dispatched.
C1, the capture's grouping key is beets' own,
`cmp.Or(AlbumArtist, Artist)` and `Album`. C2, `newShadowCmd` takes
`matcherCommit string` as a parameter and `newRootCmd` passes the
package var. C3, `internal/engine/shadow` importing no
`internal/config` stays a declared gap, guarded by the in-chain
reviewer rather than an architectural test. The wire pace ruling is
not one of C1 through C3. It was taken at the spec's own fold, ahead
of this plan, and amends the 3d handoff's recorded `DefaultInterval`
to `2 * musicbrainz.DefaultInterval`.

Nine further rulings came out of the task records themselves. Task 1,
add the TMPDIR subtest. Task 2, accept. Task 3's first round, re-point
the citation and pin the budget threshold. Task 4, accept, with a TDD
ordering deviation noted. Task 5, accept, mutation 4 unfireable and
recorded. Task 6, accept. Task 7's second round, restrict the
gate-rejected claim to review-tree-only captures. Task 8, accept
`shadowScoring` and materialize `DefaultInterval`. Task 9, accept,
mutation 5 unreachable without a signature change.

### Non-blocking leftovers, condensed

- Atomic file writes carry no directory fsync, so the crash-safety
  claim covers torn content but not the rename's own durability
  (inherited from `_salvage`, unchanged).
- The shadow store's `Observations()` and `Passes()` decode any
  non-`.json` entry a crashed `atomicfile` write might leave behind;
  filtering on the `.json` suffix would make both read paths
  crash-tolerant.
- `capture.go`'s uploader-level `os.ReadDir` tolerance (extended past
  the review's ask, to the album level alone) has no test of its own.
- `reconcile.go`'s consumed ledger only ever grows; nothing in this
  diff caps it, and at one oracle import run every fifteen minutes
  (`dubplate-import.timer`, `*:0/15`) that is roughly 35k entries a
  year.
- `reconcile.go`'s content join drops an imported row with empty
  `LibraryPaths` silently when the uploader has zero or more than one
  unclaimed capture, with no counter recording the drop.
- `match.go`'s 503 test margin (a 1s to 2s deadline) is modest on a
  loaded machine; `recordingTransport` is unsynchronized, safe only
  because requests are issued sequentially today.
- `run.go`'s `Observation.EvalDate` is the capture pass's clock while
  scoring uses the match pass's clock; nothing tests the divergent
  case because every test captures and matches in one pass.
- `run.go`'s deferral is terminal: a deferred observation carries an
  oracle signal and is never re-matched in a later pass.

### Filed for other owners

- Rung 8's `internal/engine/tagio` absorbs the shadow's provisional
  FLAC reader; `flac.go`'s own doc comment names the retirement.
- Rung 16 retires the oracle report decode (`oraclereport.go`), the
  same wind-down that retires the box-side shell import layer's other
  remnants.
- Rung 1b's registry inherits the shadow's integration class as a
  shape to absorb rather than a shape to invent, per ruling 1's own
  recommendation.

### What a later pass would be wrong to rediscover

1. `os.Rename` needs write permission on the **destination**
   directory to link the new entry, regardless of where the temp file
   was created; a chmod'd target directory fails the rename even when
   the temp file sits in `os.TempDir()`. Proving the temp-in-path
   property needs a separate `TMPDIR`-pointed-elsewhere test, not the
   crash-safety fold.
2. `os.ReadDir` already returns entries in filename order, and the
   shadow store's directory and file names are themselves the ids, so
   an explicit sort in `Observations()` or `Passes()` states a
   guarantee no test in this store can fail on removal.
3. A gate-rejected album (quarantined or rejected) never reaches
   beets and therefore never appears in a tag log; any code that
   resolves a gate-rejected capture through a backstop shared with
   beets-skipped captures must exclude tag-log matches explicitly, or
   first-match order can steal the exclusion from the wrong capture.
4. `review.Move` renames the album directory itself, so a capture
   walk can list a directory and then take `ENOENT` reading its own
   entries; that is an ordinary race to tolerate as a partial read,
   never a pass-fatal error.
5. `musicbrainz.ErrBudgetExhausted` and `musicbrainz.ErrLimiterClosed`
   are mutually exclusive within one budget: a budget that has
   already refused further spending is closed, not merely exhausted,
   and a caller branches on which sentinel it received rather than
   treating both as one retryable condition.

### 4a-ii handoff

Tasks 10 through 15 inherit the following landed surface, unchanged
unless a task's own criteria says otherwise.

- **`shadow.Config`** (`internal/engine/shadow/config.go`) and
  **`shadow.DefaultInterval`**, the wire-pace constant Task 12 wires
  into the `ClassBackground` budget alongside the 60-token cap
  (`DUBPLATE_SHADOW_BUDGET`, Task 3's registered threshold).
- **The store's layout**
  (`internal/engine/shadow/store.go`): `observations/<id>/`,
  `passes/<pass-id>.json`, `pending/<id>.json`, `resolved/<id>.json`,
  `started-at`, `consumed-reports.json`, every write through
  `atomicfile.WriteFile`.
- **`Observation`** and **`PendingRow`** (`types.go`, `run.go`), the
  engine-side shapes Task 10's marshalling into `cassette.Case` and
  `analysis.AdjudicationEntry` reads from; both stay in
  `cmd/dubplate-server` because `internal/engine/shadow` may import no
  harness package.
- **The classification's four classes**
  (`internal/engine/shadow/classify.go`): agreement (no row),
  `go-applies-oracle-skipped`, `go-skips-oracle-applied`,
  `go-applies-different-release`, plus the closed deferral set
  (`ErrUnavailable`, `ErrBudgetExhausted`, `ErrLimiterClosed`,
  `DeferUnclassified`).
- **`matchAlbum`/`shadowScoring`** (`match.go`): the accepted mutation
  site for the `FullCompat`/`QuietImportTunables` pair; `matchAlbum`'s
  signature does not change.
- **The declared gap on the config-import rule** (C3): guarded by the
  in-chain reviewer, not an architectural test. Any task that adds a
  second file importing `internal/config` from `internal/engine/shadow`
  needs the same reviewer-level scrutiny, not a free pass because no
  test enforces it.
- **The mutation sites reviewers accepted as non-literal**: Task 5's
  ReadDir-order mutation (structurally unfireable, recorded rather
  than fixed), Task 8's `shadowScoring` (the mutation site for named
  mutation 6, not `matchAlbum`'s own body), and Task 9's mutation 5
  (unreachable without a signature change).

**Resume prompt for the next session.** Task 10 is next, in the main
tree, via the pass-execute chain with this entry's common notes, after
a fresh Task 0-style staleness check against the new HEAD.

**Budget.** 4a-i spent about 4.99M: prep 2.3M, Task 0 0.07M, chain
2.14M, fix round 0.08M, simplify 0.15M, simplify review about 0.05M,
close about 0.2M (estimate). Execution against the 3M sub-ceiling was
about 2.69M, prep excluded. The plan's 5M ceiling covers all of 4a, so
4a-ii starts with roughly no headroom and needs its own ceiling before
it dispatches. That call belongs to the next conductor, not to this
close. Attended interactions for 4a-i: planning misses 1 (Task 1's
mutation-1 narrative, which a probe at planning would have caught),
execution sittings 0.

## Rung 3, pass 3d (Verdicts, the assignment solve, admission, recommendation, and the whole-corpus conformance gate): `internal/engine/tag` born, tier 1 and tier 2 grade the whole 215-case corpus (2026-09-05)

Nine plan tasks executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-05-rung-3d-verdicts-gate.md` (revision
2.1). Commits in order: `e7ff0bc` (Task 1), `e552d0a` (Task 2),
`1101a6c` plus `73af4d4` and `a78d6be` (Task 3, two fix rounds),
`19e03f7` plus `bbb65ac` (Task 4, an escalation ruled), `d89cb6f` (Task
5), `755efbd` (Task 6), `0b0db6f` (Task 7), `a69dee5` (Task 8), and
`e462210` (the close-out `code-simplifier` pass, 4 files, +61/-61).
Simplify review: accept, no blockers. The reviewer confirmed the
dropped `minD = d[final]` re-read as a no-op by stating the invariant
independently. No exported surface, contract, assertion or fixture
changed.

What landed. Task 1 built `solveLAP` in
`internal/engine/matcher/assign.go`: a hand-written dense
Jonker-Volgenant solve, structured as lap 0.5.13's own three routines,
padded to `n = max(rows, cols)` with zero, returning one entry per
column (the column-to-row orientation of fold B1). `AssignItems` builds
every cost cell with `inclArtist` false and walks the solve in ascending
track index. `ExtraItems` and `ExtraTracks` sort with
`slices.SortStableFunc`. Task 2 found and fixed a real defect in that
solver (below) and froze `testdata/lapjv-optima.json`, 404 rows mirroring
lap's own tie choice. Task 3 built `candidate.go`: `Candidates` with the
five admission gates in beets' order, insertion-order tracking, and a
stable distance sort, plus `Tunables.Validate`'s `Required` and
`Preferred` halves. Task 4 built `verdict.go` (`Recommend`'s ladder, the
`max_rec` downgrade, `ActionFor`) and `compat.go` (`Compat` as two
independent switches, `DefaultCompat`, `Apply` selecting one of four
`SD_PATTERNS` tables). Task 5 built `internal/engine/tag`, the matcher's
public entry point: `Album` ports beets' `tag_album`, with the ID-match
early return, the search fallback, and the `search_ids` branch. Task 6
added `TieEntry.Adjudicator` to `internal/harness/analysis`, required
whenever a tie row is dispositioned, refused as the `agent` sentinel on
`excluded` and accepted on `pinned`; `CheckOracleBugs` gained a
probe-case rule requiring every `oracle-wrong` hypothesis's probe cases
to resolve to an adjudication row. Task 7 built tier 1
(`internal/harness/conform/case.go`'s `GradeCases`), replaying all 215
committed cases through `tag.Album` and grading nine of
`cassette.Expected`'s eleven fields; `stub.Server` gained `Unused()`.
Task 8 wired tier 2's verdict grading into `ModeCompat` and built
`ModeDefault` as a re-score of each case's own committed candidate set
under the corrected pattern table, plus the pending-divergence path and
`CheckDefaultCompatMatchesTheLedger`.

What the gate caught. Task 1's reviewer escalated on a plan-reading
question (whether `testdata/assign/` was Task 1's or Task 2's to
create); ruled that Task 2 owns the directory and the inline Go-literal
matrices in Task 1 stand. **Task 2's mirroring test caught Task 1's
transliterated solver being genuinely suboptimal, not merely
differently tied**, on the pinned 14x14 `should-reject` matrix:
`5.283068783068783` against the recorded optimum `5.282407407407407`.
The cause was one wrong line in `findPath`'s final dual update, which
priced settled columns off the scan list's head rather than off the
augmenting path's own ending column; the two agree only while the path
ends on a column the last scan admitted. Task 1's own property test
enumerated only `n <= 6` small-integer matrices, where the defect never
fires, so the frozen 404-row table does real correctness work and must
never be thinned. No downstream figure had been derived from the
defective solver before the fix landed.
Ruling F was **not triggered**. One mirroring attempt reproduced lap
0.5.13's choice on both pinned cases and all 404 frozen rows, so no
finding document exists and `corpus/analysis/ties.json` stays
untouched by this pass.

Task 3's reviewer blocked twice. The first round found two named
mutations shipped as claimed-but-unproven. The gate-1/gate-2 admission
order is unobservable through `Add` for any input; both gates return
`(false, nil)` with no side effect, so a swap cannot be caught.
`slices.SortFunc` never reorders an all-equal-key slice at any length,
because Go's pdqsort takes a zero-swap fast path on ties, so the
three-element tie fixture could never discriminate a stable-sort
mutation. The fix added a fixture of thirteen candidates alternating
between two distance values, which does discriminate, resting on
Go's own twelve-element insertion-sort threshold (`slices.SortFunc`
diverges from `slices.SortStableFunc` starting at n=13). The gate
order itself is now held only by the `match.py:345-352` citation, with
no test able to
discriminate it; the comments were corrected to say so rather than
claim a discrimination that does not exist. The second fix round closed
a residual: a `t.Error` message still asserted the false gate-order
claim after the doc comment above it had already been corrected.

Task 4 escalated on the `max_rec` downgrade's key-resolution order. The
plan's own literal fixture keyed `tun.MaxRec` on `"nosuchpenalty"`, a
key no candidate carries, which is only detectable if `downgrade`
resolves every configured entry eagerly rather than beets' own lazy
walk over the candidate's key union (`match.py:304-322`). The conductor
ruled option (a): restore the lazy walk and change the fixture's key to
`"album"`, which the candidate fixture does carry, so the comment
becomes true as written rather than describing a divergence the port
never needed.

Task 7 escalated on scope and two undisclosed reversals. First, the
criteria's eleven-field graded-field table shipped with the verdict
pair (`Expected.Recommendation`, `Expected.Action`) unwired: `VerdictFloor`,
the verdict counters, and `casePolicy.GradeVerdict`/`VerdictFrom` were
all written and never read. Ruled: verdict grading belongs to Task 8,
which was already dispatched to consume exactly those fields. Second,
`AssignmentFloor` shipped at 214 with `TiesAwaitingRuling` empty, not
the plan's stated 212 with two pending rows; both pinned tie cases now
reproduce after Task 2's fix, so the stricter, measured 214 is accepted
and the plan's 212 figure is corrected here. Third, the tier 1 sweep
measured 59.9s against the plan's 45s trigger; ruled as a carried
finding rather than a task defect (see Forward carries below).
Mutation counts were also corrected against the real measurement; the
table below is the reconciled record. A fresh-context reviewer (opus)
re-ran the whole runner in a clean clone and accepted, reproducing
every floor, all eleven graded fields, and all five mutations
independently.

Task 8 escalated on a plan-wording mismatch (the criteria named "four
zero-instance verdict rows"; the plan's own Step 1 test code names
exactly two, `pending` and `unadjudicated`) and the same sweep-cost
finding Task 7 raised. Ruled: the plan never enumerates four rows,
Task 7's sixteen-row `policyFor` fixture set already gives every
zero-instance disposition its own firing fixture, and Task 8's two
fixtures are exactly the ones its own tests call; the "four" is a plan
wording error, not an owed deliverable. The combined two-mode sweep
costs 113.5s. That pushed `go test ./...` to 2m06s and
`scripts/check.sh` to about 3.5 minutes, both well past the plan's own
"roughly 45 seconds" trigger. Ruled that the sweep stays inside
`go test ./...`; spec `:267` remains the escape to a separate leg and
is not exercised this pass. The cost carries forward as a finding for
the first rung 3 follow-on or rung 9. Its profile is regexp
backtracking under `sdStrip` at 43% cumulative, `matcher.Levenshtein`
17.5%, `Transliterate` 7.7%.

### The verdict-path contract coverage table

Twenty contracts, V1 through V20, each with its Go site and what grades
it. A committed case grades a contract when tier 1's sweep drives
`tag.Album` or tier 2's re-score and compares the result against the
case's own record; a rule no corpus case can reach is graded by a
hand-authored fixture, with the mutation that proves the fixture fires.

| Contract | Go site | Graded by | Corpus reach |
| --- | --- | --- | --- |
| V1 normalized cost cell, `inclArtist` false | `matcher.DistanceForTrack` (`track.go:45`) via `AssignItems` | all 294 candidates' track vectors | Yes |
| V2 the solve returns a track-indexed array | `solveLAP` (`assign.go`), column-to-row orientation (fold B1) | 214 assignments; the frozen 404-row table | Yes plus fixture |
| V3 pairs in ascending track order | `AssignItems` (`assign.go`) | mutation 2, 16 cases fail (item order) | Yes |
| V4 the leftover set is one-sided | `AssignItems`'s extras | `TestAssignItemsMatchesEveryShapeOneSidedly`, all 20 shapes 0x0-3x4 | Fixture (7/14 corpus cases carry one side) |
| V5 rectangular extension is `max(rows,cols)` zero-padded | `solveLAP`'s pad step | the frozen 404-row table; mutation (pad max+1), 32 rectangular rows fail | Fixture |
| V6 extras sorted by `(disc,track,title)`/`(index,title)` | `AssignItems`'s two `slices.SortStableFunc` calls | mutation 4 (title-only sort), fires the reordered case | Fixture (total order is a corpus fact, not proven) |
| V7 five admission gates in beets' fixed order | `Candidates.Add` (`candidate.go`) | gates 3-5 fire on real cases; gates 1-2's order is unobservable through `Add` for any input, held by the `match.py:345-352` citation alone | Partial, one order pair unprovable |
| V8 `match.ignored` reads the nonzero view | `NonzeroKeys` wired at `candidate.go`'s gate 5 | hand-authored `ignored-zero-media.json` fixture | Fixture only (0 of 215) |
| V9 `match.required` tests Go zero value | `candidate.go` gate 3, thirteen-name table | `TestAddAppliesRequiredAgainstTheZeroValue` | Fixture (declared gap 3, R6) |
| V10 candidate identity is `(source,id)` | `cassette.Candidate.Source/AlbumID` | all 294 candidates | Yes |
| V11 insertion order then stable distance sort | `Candidates.Sorted` | the 13-element alternating fixture (n=13 threshold); `should-ask`/`should-reject`'s ties thinly | Thin plus fixture |
| V12 the recommendation ladder, strict-strong/inclusive-medium | `matcher.Recommend` (`verdict.go`) | mutations 1-2, `TestRecommendWalksTheLadder` | Yes, plus the hand-authored boundary ladder rows (declared gap 11) |
| V13 no-conclusion returns before the downgrade | `Recommend`'s early return | mutation 3, four committed cases (`rec-none-small-gap` and three others) | Yes |
| V14 downgrade key set is nonzero view unioned with track keys | `downgrade` (`verdict.go`) | mutations 4-5; the max_rec downgrade fires on 6 of 215 cases | Yes, thin on the track-union half (declared gap 5-6) |
| V15 `min(rec,max_rec)` over the ordered enum | `matcher.Recommendation`, `downgrade`'s lazy walk | `TestRecommendReturnsBeforeTheDowngradeOnNoConclusion` after the fix round | Yes |
| V16 the verdict-to-action mapping, quiet-first | `matcher.ActionFor` (`verdict.go`) | mutation 6 (quiet outranks timid); 215 cases, 167 apply / 48 skip | Yes |
| V17 `Quiet: true` overlaid at the grading call site | `gradeTunables`'s overlay (`conform/translate.go`) | all 215 actions; two of three `import.*` assertions are vacuous today (Timid/QuietFallback unreachable under quiet) | Yes for action, thin for the overlay's own three keys |
| V18 `tag_album`'s flow, the unsorted early return | `tag.Album` (`tag.go`) | `TestAlbumTakesTheIDMatchEarlyReturnAndIssuesNoSearch`; mutations 1-2, 4-5 | Yes, mutation 3 (sort before early return) unfirable, at most one candidate reaches that check |
| V19 `match_by_id`'s two early branches | `tag.Album`'s consensus lookup | declared gaps 2-3 (no-id, no-consensus), 0 of 215 | Fixture only |
| V20 the tracks/missing/unmatched summation order | `matcher.DistanceForAlbum`, fed `AlbumPairsFor`'s track-ordered pairs | all 294 track vectors, 55 near misses at 1-4 ULP | Yes |

### The graded-field table

Task 7's tier 1 grades nine of `cassette.Expected`'s eleven fields
(`case.go:343-436`); Task 8 adds the tenth and eleventh, the verdict
pair. Rung 15's re-run reads this table rather than deriving it again.

| Field | Grading rule | Populated by |
| --- | --- | --- |
| `Recommendation` | bit-exact string match, mode-gated by `policy.GradeVerdict` | `gradeCaseVerdict` (Task 8, `case.go:452`) |
| `Action` | bit-exact string match, same gate | `gradeCaseVerdict` (Task 8) |
| `EchoedDate` | exact string equality against the case's own `eval_date` | `gradeCase` (`case.go:408`) |
| `ConfigHash` | **not graded**: no Go analog computes a config hash | recorded non-grade |
| `Weights` | `CheckCaseWeights` against `tun.DistanceWeights` | `gradeCase` (`case.go:411`) |
| `Thresholds` | `CheckCaseThresholds` against `tun` | `gradeCase` (`case.go:414`) |
| `Candidates` | identity, distance (epsilon or bit-identical per policy), penalty view, per-track vector | `gradeCandidates` (`case.go:706`) |
| `Assignment` | item-to-track pairs against `AssignItems`'s solve, gated by `policy.GradeAssignment` | `gradeAssignment` (`case.go:783`) |
| `ExtraItems` | exact index-list match | `gradeAssignment` |
| `ExtraTracks` | exact index-list match | `gradeAssignment` |
| `AppliedFields` | **not graded**: no Go analog computes post-apply in-memory item fields | recorded non-grade |

### The declared corpus-coverage gaps

**Eleven new gaps, taking rung 3's running total to twenty-one** (3a's
three, 3b's six, 3c's one, this pass's eleven). Rung 4 and rung 15
should read this total rather than re-count it; the plan's own prose
states "ten new" in one place but enumerates eleven items and the
running-total arithmetic (10 + 11 = 21) requires eleven, so eleven is
what this entry records.

1. `import.search_ids` and the explicit-ID branch (0 of 215).
2. `match_by_id`'s no-id branch (0 of 215).
3. `match_by_id`'s no-consensus branch (0 of 215).
4. `import.timid`'s gate on the ID-match early return (0 of 215).
5. The `max_rec` downgrade read against a key added at 0.0 (0 of 294
   candidates carry `missing_tracks`/`unmatched_tracks` at exactly
   0.0, though 272 carry some key at 0.0).
6. `_recommendation`'s per-track key union in the downgrade set: the
   loop runs and never changes an outcome under the shipped `max_rec`.
7. `import.quiet = false` and the whole `timid`/`none_rec_action`/"ask"
   half of `_summary_judgment` (the shim forces quiet).
8. `quiet_fallback: asis` and the `ASIS` action entirely.
9. Every non-default value of the four recommendation knobs: all 215
   cases echo the shipped defaults, so the tier 1 cross-check proves
   the transcription, never the behavior.
10. The assignment objective's `inclArtist=false` on a VA release: no
    tied or assignment-sensitive VA case exists.
11. The recommendation ladder's strict-versus-inclusive asymmetry: no
    committed candidate sits exactly on a threshold (a sweep over all
    215 cases finds none at 0.04 or 0.25); the hand-authored ladder
    rows are the sole detector.

Three items are explicitly not counted. `match.required` and
`match.ignored` are 3a's gaps 1 and 2 already; 3d builds their
consumers and a fixture for each, which closes them by unit test rather
than adding new gaps. "`strong` reached after a text search" is a
corpus fact, not a code path this pass builds. `profile_hashes`
covering one profile of three is a corpus-lane item open since 2b,
carried below rather than counted here.

**Two surfaces are non-ports, not gaps.** `bug-assignment-rejection-gap`'s
fix needs a per-pair rejection mechanism beets does not have.
`bug-nonlatin-alias-selection`'s fix needs a corrected applied artist
string, which rung 3 computes none of. Both join the improvements
register.

### The mutation table, measured counts

Several plan-stated mutation counts were corrected against the real
measurement.

| Task | Mutation | Plan's figure | Measured |
| --- | --- | --- | --- |
| 3 | swap gate 1 (zero-track) and gate 2 (duplicate) | fires | **unobservable**: both gates return `(false, nil)` with no side effect, for any input |
| 3 | `slices.SortFunc` for `slices.SortStableFunc`, 3-candidate tie | fires | **does not fire**: pdqsort takes a zero-swap path on an all-equal-key slice at any length |
| 3 | same mutation, 13-candidate alternating-distance fixture | (not in plan) | **fires**, resting on Go's 12-element insertion-sort threshold |
| 7 | pairs in item order | 23 cases | **16 cases** (uniform per-track distances hide the permutation on the `adv-assign-*` family) |
| 7 | insertion order rather than distance order | 21 cases | **18 cases**, 60 candidate-identity mismatches |
| 7 | truncate the search fan-out | (unspecified) | **48 cases fail with `Unused()` in place, 22 without** |

### Evidence

The three earlier tiers are unchanged. Tier 0 still replays 72/72
bit-identical, translation still grades 50/50, and the request tier
still runs 215/303/48 at zero misses. Tier 1
(`GradeCases`, `ModeCompat`): 215 cases, 294 candidates, 294 album
distances, 294 track vectors, 214 assignments compared, 55 near misses
(1-4 ULP, max relative drift `4.61e-16`), zero failing. Tier 2:
`VerdictFloor` 215; `ModeDefault`'s ten `defect_free` rows reproduce
bit-identically; `ModeDefault` moved zero non-probe verdicts (see
`docs/superpowers/research/2026-09-05-rung-3d-default-mode-verdict-moves.md`);
eight probe cases `Ungraded` (`ProbeCaseFloor` 20, minus the ten
`defect_free` rows, minus the two `ProbeCaseOracleCorrectRows`);
`TiesExcluded` 0 against its ceiling. The corpus totals were
independently recomputed off the tree and matched every floor (215
cases, 294 candidates with 272 carrying a 0.0-valued key, 23 non-identity
assignments, 21 multi-candidate cases, 7/14 non-empty extras).

Measured wall times: tier 1 sweep 60.3s (70.3s package) at `-count=1`;
tier 2's combined two-mode sweep 113.5s (`ModeCompat` 60.3s shared via
`sync.OnceValues`, `ModeDefault` 53.2s); `internal/harness/conform`
123.7s; `go test ./...` 2m06s; `bash scripts/check.sh` about 3.5
minutes.

### The rung 3 to rung 4 handoff

What rung 4's spec needs from 3d, ruling B option (b): rung 4 deploys
at functional completeness with every divergence stamped with the
matcher commit.

1. **The entry point.** `tag.Album(ctx, src *musicbrainz.Client, items
   []matcher.ItemFields, tun matcher.Tunables, compat matcher.Compat,
   evalDate time.Time, opts tag.Options) (*tag.Proposal, error)`, in
   `internal/engine/tag`. `ctx` first, `evalDate` a parameter and never
   a clock read, `opts.SearchIDs` the only optional input. `Proposal`
   carries `CurArtist`, `CurAlbum`, `Candidates` (ascending distance,
   stable), `Recommendation` and `Action`.
2. **The divergence stamp.** Every shadow divergence row rung 4 writes
   carries the matcher commit, so a divergence from a superseded
   matcher revision is re-evaluated mechanically rather than consuming
   owner adjudication. It joins `analysis.AdjudicationEntry` beside the
   existing fields; the shadow's subject shape (`{album, release_mbid}`,
   `adjudication.go:35`) is already modeled and carries zero committed
   rows, so rung 4 is its first producer.
3. **The shadow's budget class.** `musicbrainz.NewClient` requires a
   `Budget` and every request `tag.Album` issues spends one token, so
   rung 4 caps its own MusicBrainz spend by the budget it builds the
   client with. The shadow's class is `ClassBackground`, deferred
   behind rung 7's user-present probe, and the wire pace is
   `shadow.DefaultInterval = 2 * musicbrainz.DefaultInterval`
   (corrected 2026-09-07 at the 4a-ii close; this handoff recorded the
   limiter's own `DefaultInterval = time.Second`, 3c contract C20,
   which the conductor's pace ruling at rung 4's spec fold doubles for
   the shadow's `ClassBackground` caller). Rung 4 branches
   on `musicbrainz.ErrUnavailable` to defer rather than record a
   divergence, and on `tag.ErrNoCandidates` to record "no match".
4. **The pending-divergence file format.** A row a human commits into
   `corpus/analysis/adjudications.json` carries `id`, `subject.case_id`
   (or, for the shadow, `subject.album` plus `subject.release_mbid`),
   `suspected`, `citation`, `release_truth`,
   `disposition: pending-adjudication`, `reason`, `adjudicator: agent`
   and `date`, all five prose fields non-empty. The runner prints
   exactly that row and writes nothing; rung 4's shadow inherits the
   same rule, and the launch gate refuses any pending row.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, unchanged, travels forward from every
   rung 3 pass close (quoted in 3c's entry above).
2. **The four-knob wall registration** stays the named follow-on task
   for the first rung 3 pass after rung 1b lands; `AdminEmail`'s
   interim User-Agent contact sits beside it, both unmoved this pass.
3. **`ties.json` measures `Candidates[0]` only**, so 80 of the 294
   committed candidate assignments were never asked the tie question,
   and **17 of them are tied with no disposition** (re-measured by
   Task 7's own sweep, the first run able to solve every candidate),
   across `real-noise-dylan-bootleg`, `real-noise-pink-floyd-de1994`,
   `real-nonlatin-girls-generation`, `real-nonlatin-utada-first-love`,
   `real-nonlatin-utada-first-love-locale`, and the non-best candidates
   of `should-ask` and `should-reject`. Five change tier 1's graded
   `Tracks` vector on an alternative optimum and two also change the
   `tracks` penalty sum. **Filed for lane E1** as a derivation ask;
   tier 1 grades all 294 vectors meanwhile.
4. **The `lapjv` extension model is behavioral, not source.**
   `n = max(rows, cols)` with zero fill, established against `lap`
   0.5.13 on 2634 random rectangular matrices and reconfirmed on 262 at
   this plan's drafting; the classic `(rows+cols, max+1)` recipe
   `derive_ties.py`'s own docstring assumes disagrees on 30% of
   rectangular ties. Frozen as
   `internal/engine/matcher/testdata/lapjv-optima.json` with its
   provenance header. The day `corpus/manifest.json`'s `lap` pin moves,
   that table must be regenerated, and the oracle-tagged differential
   test this pass declined becomes worth building.
5. **`MergePseudoRelease`'s aliasing hazard is now rung 9's alone.**
   `AlbumForID` closes it by construction because only the merged value
   escapes, and `tag.Album` inherits that same call site unchanged. The
   `slices.Clone` is owed the day a caller retains both values, which is
   rung 9's pipeline.
6. **`vaLikely` stays unexported, permanently.** `match_by_id`'s
   no-consensus branch is corpus-unreachable and the search path
   computes `vaLikely` inside `AlbumSearchFor`, so no caller outside
   `musicbrainz` ever needs it. 3c's carry 13c is closed as "stays
   unexported", recorded so a later pass does not export it for
   symmetry.
7. **The retry path is ungraded by the corpus and always will be**: the
   stub answers 200 or 404 only, `cassette.Request` carries no status
   field, and the response-status extension stays declined, now for the
   fourth pass running.
8. **Three regime divergences travel together**, unchanged from 3c:
   3b's `\b` in `SDPatternsWordBoundary` and now in
   `SDPatternsGuardedNoBracketDiscount`, the query path's
   full-versus-simple case lowering, and `_extract_id`'s `\w`. All three
   are Unicode-aware in Python and ASCII-only in Go RE2, and all three
   are benign on the committed corpus.
9. **`Case.Extra` still has no producer** (3c guard 5), zero of 215
   cases carry an entry, and `Unused()` deliberately excludes `Extra`
   because a strong ID match legally orphans search captures into it.
10. **`profile_hashes` covers `locale` alone** (open since 2b) and the
    response-status extension stays declined (open since 2c). Neither
    moved this pass.
11. **The measured wall times**, listed in Evidence above, join the
    version-bump differential doc the same courtesy every rung 3 pass
    close has passed forward.
12. **The tier 1/tier 2 sweep cost** (60.3s/53.2s, conform 123.7s,
    `go test ./...` 2m06s) is carried, with its CPU profile (regexp
    backtracking under `sdStrip` 43% cumulative, `matcher.Levenshtein`
    17.5%, `Transliterate` 7.7%), as the first rung 3 follow-on or
    rung 9's, and `scripts/check.sh` gains no separate leg for it this
    pass (spec `:267` stays the future escape).
13. `VerdictFrom` is gated on mode; a mode-blind read fired on 18 real
    cases in an intermediate build and was corrected before Task 8's
    close.
14. `nearMissTolerance` (1e-12 relative) is about 4500 ULP against the
    4 ULP maximum drift this pass measured; the tolerance is spec-
    mandated (`:141`), not a hole.
15. `cassette`'s two `Request` rows sharing `(Method, PathQuery)` are
    not rejected, and `Unused()` would mark both hit on one issue;
    zero committed cases are affected, pre-existing from 3c.
16. The `ties skipped` policy arm (`case.go:585-590`) reads wider than
    the spec row's own sentence (it also sets `ExpectNoCandidate`);
    harmless, cannot ungrade anything real.
17. Two of the three `import.*` overlay assertions in
    `translate_test.go` are vacuous today (`QuietFallback` and `Timid`
    equal their shipped defaults); only `NoneRecAction` discriminates.
18. The plan's fixture-count wording ("four zero-instance verdict
    rows") is never enumerated anywhere in the plan; ruled a wording
    error, not an owed deliverable (Task 8's escalation).
19. The 17 unledgered tied candidate assignments (item 3 above) are
    filed for lane E1 as a `derive_ties.py` extension ask (R19).
20. **Owner ask, batched rather than raised** (Resolution R1):
    re-dispositioning `bug-assignment-rejection-gap` from `oracle-wrong`
    to `no-correct-answer`, whose semantics the spec's own grading
    table already describes as exactly the behavior this pass
    implements by hand. `disposition` is owner-only and this pass
    writes none.

### Filed for other owners

- **The oracle-tagged `lapjv` differential test**, for the lane that
  owns `internal/harness/oracle`, to be built the day the `lap` pin
  moves (item 4 above).
- **Extending `derive_ties.py` to ask the tie question of every
  candidate rather than `Candidates[0]`**, for lane E1 (item 3 above).

### Non-blocking leftovers, condensed

- `internal/engine/matcher/assign.go`'s dual-update fix left one
  comment reading as though the augmenting path's ending column carries
  information the running minimum does not; harmless, but worth a
  clearer sentence the next time that function is touched.
- `internal/harness/analysis/ties.go:140-145`: a unique (untied) row
  carrying a stray `Adjudicator` with no `Disposition` is accepted
  silently, unlike the symmetric skipped-row check; not required by the
  criteria.
- `internal/harness/conform/case.go:495`: `compareFloat` has no direct
  unit test; its epsilon path rides on the corpus's 55 near misses
  alone, and the flagged-row bit-identity refusal is proven by mutation
  only, not by an isolated table test.
- `internal/harness/conform/case.go:602`: `gradeCaseDefault` sets
  `res.graded = true`, a field `GradeCases` never reads on the
  `ModeDefault` path; dead but harmless.

### Ruled

**Revision 2.1's three conductor rulings were all taken as
recommended.**

1. B2, `ModeDefault` re-scores each case's own committed candidate set
   under the corrected table and issues no search; the request-set rule
   stays `ModeCompat`-only.
2. m16, the entry point keeps the package name `internal/engine/tag`,
   with both doc comments naming the `tagio` pairing, and the build
   spec's tree listing gained the distinguishing clause this pass
   drafted.
3. m15, the `docs/` ledgers row is written into the parallel-approach
   spec's contested-surface table.

**Every execution-time ruling, in dispatch order.** Task 1's escalation
(fixture location): Task 2 owns `testdata/assign/`; the inline
Go-literal matrices in Task 1 stand. Task 3's fix round (twice): both
findings accepted and fixed, no further escalation. Task 4's escalation
(the `max_rec` key-resolution order): option (a), restore beets' lazy
walk, fixture key `"nosuchpenalty"` to `"album"`. Task 7's escalation
(verdict fields, `AssignmentFloor`, sweep cost): verdict grading is
Task 8's; `AssignmentFloor = 214` with an empty `TiesAwaitingRuling` is
accepted as the measured truth over the plan's stated 212; the sweep
cost is carried as a finding. Task 8's escalation (fixture count,
sweep cost): the plan's "four zero-instance rows" is a wording error,
two fixtures stand; the combined sweep stays inside `go test ./...`.

### Carried

- **To 3d, now resolved.** `Preferred.Media`/`Preferred.Countries`
  config-load validation is **CLOSED** (Task 3): `Tunables.Validate`
  compiles both lists and returns a wrapped `ErrInvalidTunable`. The
  compat switch's pattern-table selection is **CLOSED** (Task 4,
  Resolution R2): both hypotheses default off, proven by the fourth
  `SDPatternsGuardedNoBracketDiscount` table and `ModeDefault`'s ten
  `defect_free` rows. Tie mirroring needing both `inclArtist` call
  shapes is **CLOSED** (Tasks 1-2). 3c's R1 seam (the ID-match early
  return, `_recommendation`, `_add_candidate`, the explicit
  `search_ids` branch) is **closed for the first three, declared as
  gaps 1-3 for the last two**. R6's stub blind side is **CLOSED**
  (Task 7, `stub.Unused()`, a hard failure in tier 1). `vaLikely`'s
  export question is **CLOSED as "stays unexported"** (item 6 above).
  `MergePseudoRelease`'s aliasing hazard is **carried to rung 9** (item
  5 above). R7's `Case.Extra` producer, R11's `AdminEmail` interim
  contact, the four-knob wall registration, the retry path, and
  `profile_hashes` are all **carried unchanged** (items 2, 7, 9-10
  above).
- Five 3b `code-simplifier` leftovers still sit in HISTORY, none
  blocking, none touched this pass.
- Adjudication batch 1 stays fully ruled, unaffected by this pass; no
  new adjudication row was produced (ruling F never fired).

Budgets: prep 2.51M, Task 0 0.09M, run A 0.32M, run B 1.34M, fix round
0.07M, runner review 0.14M, run C 0.32M, simplify 0.15M, simplify
review about 0.07M, close ≈0.15M (estimate), summing to about 5.15M
against the 6M ceiling re-derived after Task 2 to an expected 4.8M and
not raised. Attended interaction points: planning misses 2 (Task 1's
`testdata/assign/` fixture location and Task 8's "four zero-instance
rows" fixture count, both ambiguities a planning question would have
caught); execution sittings 0, every fork ruled by the conductor under
the 2026-09-05 grant, ruling F and the pending-divergence path never
firing.

## Rung 3, pass 3c (The client, the limiter, and request-construction grading): `internal/engine/musicbrainz`'s `query.go`, `client.go` and `limiter.go` born, 351 requests graded byte for byte (2026-09-05)

Eight plan tasks executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-04-rung-3c-client-limiter.md`
(revision 2.2). Commits in order: `e9364ff` plus `e90bdad` (Task 1),
`8e7dc21` (Task 2), `f3c3edf` plus `208c5af` (Task 3), `4e7bb04`
(Task 4), `807dc45` (Task 5, escalated), `f088b1a` (the conductor's
plan revision 2.2 ruling), `dca3d17` plus `18ea5da` (Task 6), `18a1292`
(Task 7), `78a6e5a` plus `0082ae6` (Task 5b, the escalation's fix, two
commits closing the fourth seam), `7809558` (the close-out
`code-simplifier` pass, 9 files, +47/-38), and this entry (Task 8).

What landed. Task 1 reshaped `cassette.Case.Extra` from `[]string` to
`[]Request` in `internal/harness/cassette/cassette.go`. The four
per-entry checks the Requests loop already applies now cover it, plus
three exclusivity rules keyed through a new unexported `requestKey`.
The task bumped `wantManifestSchemaVersion` and
`corpus/manifest.json`'s `schema_version` from `"3"` to `"4"`, and
added `TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema`,
proving the migration empty over all 215 committed cases. Task 2 gave
`internal/harness/stub` a second replay map: `extra
map[key]extraEntry`, consulted after the request map. `Consumed()
[]cassette.Request` reports distinct from `Misses()`, and the
derivation-inertness rule is stated on both doc comments (zero of 215
cases carry an `Extra` key today). Task 3 added `ItemFields.Comp` and
`boolField`, and exported `IsVariousArtistsName`. It closed 3a's
`dataSourcePenalty` carry with `defaultDataSourceMismatchPenalty =
0.5`, replacing the tunable fallback and removing both `conform`
compensations. It also corrected the
`SearchLimit`/`DataSourceMismatchPenalty`/`DataSourcePenalties`/
`AddDataSource` doc comments. Task 4 built `query.go`: the sixteen-
element `releaseIncludes`, unexported `vaLikely`, and `albumCriteria`.
`albumCriteria` walks `tun.ExtraTags` with the
`tracks`/`alias`/`catalognum`/`date:(0)` value rules (the tenth
declared corpus-coverage gap). The rendering layer (`escapeLucene`,
`formatSearchTerm`, `lowerSearchTerm`, `trimSearchTerm`) reproduces
beets' double lower-and-strip and its escape-before-wrap order. Task 5
built `limiter.go`: the three-class dispatcher with
`reservationInterval = 4` and a round-robin reserved token over the
non-empty classes below the one served. It also added `Budget`'s
reserve-then-release cap, `Unmetered()`, and `Close()`.
`NewLimiter(interval time.Duration)` carries `DefaultInterval =
time.Second` and C20's citation chain. Task 5b closed the escalation
(below). Task 6 built `client.go`: `ClientConfig` and `NewClient`,
requiring a `UserAgent` and a `Budget`. `Release` and `SearchReleases`
force `fmt=json` through `url.Values.Encode`, backed by
`extractReleaseID`. The seven-attempt retry transport uses urllib3's
backoff schedule and both `Retry-After` forms, with a per-attempt
limiter token. `AlbumForID`'s pseudo double fetch and merge closes 3b's
`MergePseudoRelease` aliasing carry for this pass's own call site.
Task 7 added `internal/harness/conform/request.go`. `GradeRequests`
opens one stub per case. It issues one `client.Release` per distinct
logged release mbid, and one `client.SearchReleases` when the case
logs a search. It joins every `srv.Misses()` entry into the case's
error. It asserts the three corpus floors (215 cases, 303 lookups, 48
searches) plus a per-case issued-count check that a silent skip cannot
pass.

What the gate caught. `bash scripts/check.sh` stayed green at every
task end. Three tasks took a fix round from the in-chain
`diff-reviewer`, and Task 5 escalated past a fix into a plan
correction. Task 1's reviewer found only five of the plan's six named
Extra-loop failure proofs landed, with the method-non-empty and
response-non-empty per-entry checks unfired by any test. The fix
commit `e90bdad` added both, each verified by a targeted mutation.
Task 3's reviewer blocked because the conductor's ruling (fold M-a)
required `DefaultTunables` to build `DataSourcePenalties["MusicBrainz"]`
from the `DataSourceMismatchPenalty` field. The first attempt instead
shared one local variable between both, leaving the field with zero
production readers despite its doc comment's claim. `208c5af` rewrote
`DefaultTunables` to build the map from the field directly, so a
mutation of the field's literal now provably moves the map too. Task
6's reviewer blocked on an undelivered behavior-contract bullet: the
deliberate divergence on an absent release `status` key (a plain
non-pseudo release here, a dropped candidate via beets' `KeyError` in
`beetsplug/musicbrainz.py:785`) was nowhere in the tree. `18ea5da`
recorded it as a comment at the `PseudoReleaseTarget` call site with
the 303-of-303 corpus measurement that makes it unreachable today, and
a matching doc comment on `Release.Status`. Task 5 escalated rather
than blocked. The plan's cancelled-waiter test cancelled and joined
the departing waiter before feeding a tick, so `depart()` had already
dequeued it. `serve`'s refused-offer arms (`case <-w.gone:`, `case
<-l.done:`) carried zero coverage. That left unproven the
distinguishing half of the chosen cancellation shape, which tries the
next waiter rather than parking the token. The conductor ruled option
(a) at plan revision 2.2: a fourth unexported hook, `holdBeforeOffer`.
It lets a test block `serve` between `pick()` and the offer, cancel
the picked waiter there, and assert the next waiter takes that same
tick's token. Task 5b landed the fix in two commits: `78a6e5a` (the
in-flight refusal, both mutations of the refusal path killed) and
`0082ae6` (the shutdown arm, `case <-l.done:`). That arm had never
been exercised either; deleting it hangs the test to its timeout. The
seam count is **four** (`newLimiter`, `awaitQueued`, `awaitTicks`,
`holdBeforeOffer`), not the plan's originally named two. A final
`diff-reviewer` (opus) read `7809558` and `0082ae6` together against a
fresh checkout and returned accept, confirming the seam count, the
`-race -count=20` limiter replay, and both new assertions' mutation
proofs.

### The request coverage table

Twenty contracts, C1 through C20, each with its Go site and what
grades it. A committed case grades a contract when `GradeRequests`
(Task 7) issues a real request through it and compares the resulting
bytes against the case's own log; a rule no corpus case can reach is
graded by a hand-authored fixture in the owning package's own
`testdata`, with the mutation that proves the fixture fires.

| Contract | Go site | Graded by | Corpus reach |
| --- | --- | --- | --- |
| C1 `fmt=json` forced, no exception | `client.go` `releaseRequestURL`/`searchRequestURL` via `url.Values.Encode` | all 351 requests, Task 7's sweep | Yes |
| C2 API root `<base>/ws/2`, three config keys ignored | `client.go` `wsPathPrefix`, `ClientConfig.BaseURL` | the injected `BaseURL` is the only root the Go client models; the three ignored keys (`https`, `ratelimit`, `ratelimit_interval`) have no Go analog to grade, recorded as a non-membership | n/a, structural |
| C3 sixteen-element `RELEASE_INCLUDES`, one `inc=` value | `query.go` `releaseIncludes` | all 303 lookups, one distinct `inc=`; `TestReleaseIncludesJoinsAllSixteen` (hand literal) plus mutation 1 (drop `genres`, 303 misses measured) | Yes |
| C4 a search carries no `inc` | `client.go` `searchRequestURL` | all 48 searches carry none | Yes |
| C5 the nineteen-character Lucene escape set | `query.go` `escapeLucene` | `TestEscapeLucene`'s 30-row table; mutation 2 (drop `-`) fires 6 cases | Thin: `%27`/`%2C` unescaped and `%5C` escaped both appear in the corpus, but most of the nineteen ride on the fixture table alone |
| C6 lower, strip, drop-if-empty, escape, then wrap | `query.go` `formatSearchTerm` | end-to-end query rows verified against committed `path_query`; mutation 5 (escape after wrap) fires both wrapper rows | Yes |
| C7 space-joined terms, empty terms filtered | `query.go` `AlbumSearchFor` | `TestAlbumSearchForDropsEmptyTerms` (hand fixture, the plan's extra row) | Thin: no committed case has an empty album name |
| C8 insertion-ordered criteria, eight-key `FIELDS_TO_MB_KEYS`, three value rules | `query.go` `albumCriteria`, `fieldsToMBKeys` | `extra_tags.json` (the tenth declared gap); the corpus-reachable `release`/`artist`/`arid` rows | Partial: the base three criteria are reached, the eight extra-tag keys are not (declared gap) |
| C9 the pre-filter's second lower-and-strip | `query.go` (folded into `formatSearchTerm`'s own lower/strip; C9's dict comprehension has no separate Go site since the tuple's query-text half is always `""`) | the double-lower-and-strip end-to-end rows | Yes, jointly with C6 |
| C10 `va_likely`'s three-way disjunction, the `search_artist` override | `query.go` `vaLikely` | mutation 6 (drop the `comp` disjunct) fires the `comp` row and the arid row | Yes |
| C11 `search_limit` 5, legacy spelling `musicbrainz.match.searchlimit` | `query.go` (`AlbumSearch.Limit = tun.SearchLimit`); `tunables.go`'s corrected doc comment | all 48 searches carry `limit=5` | Yes, pins the default only |
| C12 the transport policy (retry count, backoff schedule, timeout, `Retry-After`) | `client.go` `doWithRetry`, `retryWait`, `backoff` | `TestBackoffSchedule`'s six-row table; `TestRetryAfterDeltaSecondsHonored`; mutations 5-6b | Yes, through hand-authored fixtures; no committed case's stub answers a non-200/404 status |
| C13 fetch, translate, conditional second fetch | `client.go` `AlbumForID`, `PseudoReleaseTarget` | mutation 9 (skip the pseudo second fetch, `Country` assertion fails) | Yes, the 3 pseudo cases |
| C14 per-id fan-out follows search-response order | `client.go` `SearchReleases` (`[]string` in response order) | `TestSearchReleasesDecodesTheObjectShapeAndPreservesOrder` | Yes, implicit in every multi-hit search |
| C15 a 404 drops a lookup candidate silently, propagates from search as an empty list | `client.go` `Release`/`SearchReleases` mapping to `ErrNotFound` | the zero-miss rule (Task 7); no test asserts the 404-to-`ErrNotFound` mapping directly (recorded gap, not a plan mutation) | No committed case's stub answers 404 |
| C16 `search_query_ascii` is a structural no-op | `doc.go`'s stated non-check | nothing to grade; the absence is the contract | n/a |
| C17 the corpus's request census (215/303/48/167) | `internal/harness/conform/request.go` `RequestCaseFloor`/`RequestLookupFloor`/`RequestSearchFloor` | `TestRequestConstructionGradesCleanly`, exactly matching the plan's measured figures | Yes |
| C18 `net/url` reproduces the canonical form byte for byte | `client.go`'s `url.Values`-built URLs | all 351 requests' `path_query` compared byte for byte | Yes |
| C19 a search response is an object, ids under `releases` | `client.go` `SearchReleases`'s decode struct | all 48 searches; mutation 4 (bare-array decode) fails all 48 | Yes |
| C20 the wire pace is 1/s, not 4/s | `limiter.go` `DefaultInterval = time.Second` | `TestNewLimiterPacesFromItsOwnTicker`; no corpus case measures wall-clock pacing (Unmetered budget in Task 7's sweep) | n/a, a live-spender fact rather than a byte comparison |

**The one new declared corpus-coverage gap: `albumCriteria`'s
`extra_tags` branch**, including `tracks`, `alias`, the `catalognum`
space-strip and the `date:(0)` trap on an unset year, unreachable
because no committed profile writes a `musicbrainz:` block setting
`extra_tags`. **Together with 3a's three and 3b's six, rung 3 now
carries ten declared gaps**; 3d's close should read that running total
rather than re-count it.

**Four surfaces are non-ports, recorded with their reasons rather than
as gaps.**

1. The recording lookup and `RECORDING_INCLUDES`. This pass ships only
   the release lookup and release search beets' album path uses.
2. The browse fallback (`_ensure_complete_recordings`). Carried
   forward from 3b, still unported.
3. Beets' explicit `search_artist`/`search_name` override on a
   user-forced search. No committed profile sets it, and the plan
   scopes 3c to the automatic album-match path.
4. urllib3's 413-plus-`Retry-After` sixth retryable case (`fold m3`).
   Unreachable from MusicBrainz, which never answers 413.

### Evidence

The request tier, `TestRequestConstructionGradesCleanly`: **215 cases
graded, 303 lookups, 48 searches, 167 cases with no search**, zero
misses, in 3.34s at `-count=1` (measured at close; the task chain's own
figure was 3.32-3.54s across runs). `Consumed()` accumulates zero
entries across the sweep, matching C17's zero committed `Extra` keys.
The schema bump's round-trip proof,
`TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema`, ran 1.870s
at `-count=1` over all 215 cases at close (2.050s to 2.662s across the
task chain), confirming no committed byte moved. Tier 0:
`TestComponentTableReplaysAllRows` still 72/72 rows, bit-identical, in
0.015s at close, per-method counts unchanged after the
`dataSourcePenalty` fix and the removal of both `conform`
compensations (add 8, add_data_source 4, add_equality 3, add_expr 2,
add_number 4, add_priority 4, add_ratio 4, add_string 3,
generic_penalty_keys 1, string_dist 23, track_distance 16). `go test
./...` ran 15.081s wall at close (`internal/harness` 11.255s and
`internal/harness/conform` 11.230s dominating, tier 0 and the request
tier both running inside the latter). `bash scripts/check.sh` ran
1m25.932s wall at close, gitleaks 1m4.3s of that scanning ~114.4 MB,
consistent with the task chain's 1m6.7s-3m00.7s range across tasks.
The limiter's `-race` run: `go test -race -count=1
./internal/engine/musicbrainz/...` 2.090s at close (1.162s measured
under Task 5 alone); the pass-end reviewer additionally ran the ten
limiter tests under `-race -count=20` with zero failures, and the
Task 5b review ran `-race -count=1` over both
`internal/engine/musicbrainz` and `internal/harness` (101s and 123s)
plus a clean `golangci-lint` over `internal/...`.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, verbatim from the plan's Spec
   section, travels forward from every rung 3 pass close: "Rung 15
   re-runs `internal/harness/conform`'s three tiers over the complete
   corpus, at the matcher commit as landed, and its acceptance is the
   launch bar's matcher clause read against that run: zero false
   confident matches over every stratum including stock-config, and
   zero unadjudicated beets-right/Go-wrong cases. Rung 3's own
   acceptance is provisional against the corpus at its then-current
   size. Any tier-2 divergence carrying a pending row at rung 15 is the
   launch gate's business, not rung 15's to wave through."
2. **The four-knob wall registration** stays the named follow-on task
   for the first rung 3 pass that runs after rung 1b closes:
   `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, `SearchLimit`,
   `internal/config/keys.go` the file no lane edits until then, and
   `Threshold.Default` is a string with `ParseInt` the only parser, so
   the three float rec thresholds need a float variant the registry
   does not yet have.
3. **3c grades request construction, not the request set.** The seam
   is R1's: 3d owns the ID-match early return, `_recommendation`,
   `_add_candidate`, the explicit `search_ids` branch, and
   `match_by_id`'s no-id/no-consensus branch, and inherits the stub,
   the zero-miss rule, and the client unchanged.
4. **The stub's blind side and its four-line close** (R6). Nothing
   reports which logged `Request` entries a run consumed, so a client
   issuing fewer requests than the log records leaves zero misses.
   3c's per-case issued-count assertion covers it inside 3c; 3d needs
   the hit set.
5. **`Case.Extra` has no producer** (R7), and a producer is
   `cmd/harness` work in lane E1's territory. The channel stays
   hand-authored until E1 builds one.
6. **The engine's `/ws/2` constants are a deliberate second copy**
   (R9), and unifying them with `internal/harness/mbrequest` would turn
   the request grading into a self-derived expectation. 3b's "worth
   revisiting if a fourth appears" trigger is met and answered.
7. **The client's three deliberate transport divergences** (R13): a
   30s timeout rather than beets' 10s; a limiter token acquired per
   attempt rather than per call; and `Retry-After` honored in its
   delta-seconds form only, the HTTP-date form falling through to the
   computed backoff because parsing it needs a clock read
   `TestEngineNeverReadsTheClock` forbids. Beside them, one non-port:
   urllib3's 413-plus-`Retry-After` sixth retryable case, unreachable
   from MusicBrainz. None is gradeable by the corpus.
7b. **The wire pace is one request per second and it is now written
    down** (C20): beets' `musicbrainz.org` branch pins `rate_limit =
    1.0` and its session limiter binds before the 0.25s adapter floor,
    `musicbrainz.DefaultInterval` is `time.Second`, and rung 4's shadow
    is the first real spender that inherits it.
7c. **The `AdminEmail` User-Agent contact is interim** (R11).
    `Config.AdminEmail`'s own doc comment says adopters compare it and
    never display it, and the harness precedent this plan cites uses a
    dedicated `DUBPLATE_HARNESS_CONTACT` input instead. A dedicated
    contact row is the post-1b follow-on, beside the four-knob
    registration.
8. **The retry path is ungraded by the corpus and always will be**:
   the stub answers 200 or 404 only, `cassette.Request` carries no
   status field, and the response-status extension stays declined,
   now for the third pass running.
9. **Three regime divergences now travel together** (R14): 3b's `\b`
   (Unicode-aware in Python `re`, ASCII-only in Go RE2), the query
   path's full-case-versus-simple-case lowering with all 56 sweep code
   points reaching the wire, and `_extract_id`'s `\w`. The strip
   divergence is closed rather than carried.
10. **`AlbumForID` closes 3b's `MergePseudoRelease` aliasing carry by
    construction**, and the hazard changes owner to 3d and rung 9
    rather than disappearing; the `slices.Clone` is the fix the day a
    caller retains both.
11. **3a's `dataSourcePenalty` carry is CLOSED** (R15), with both
    `conform` compensations removed and 72/72 tier-0 rows re-verified
    as the proof.
12. **An absent release `status` is a plain non-pseudo release here
    and a dropped candidate in beets**, unreachable in the corpus (303
    of 303 lookups carry one), recorded as a deliberate choice at
    `client.go`'s `PseudoReleaseTarget` call site.
13. **The limiter's one knob is `reservationInterval`**, and rung 9
    re-verifies that integer against its real callers rather than
    re-deriving the policy. Rung 9 also inherits the reserved token's
    **round-robin over the classes below the served one**, which
    replaced revision 1's "lowest non-empty" rule because that rule
    starved `ClassBatch` outright whenever a probe stream and the
    shadow were both queued.
13b. **`Budget`'s cap reserves before it waits**, so it holds under
     concurrent callers; a pre-wait check is a lost update that no
     `-race` run reports.
13c. **`vaLikely` is unexported at rung 3** and 3d exports it the day
     `match_by_id`'s no-consensus branch needs it; the same rule kept
     `stub`'s `Unused()` accessor out of this pass.
14. **The measured wall times**: the request tier 3.34s, the
    round-trip proof 1.87s, tier 0 0.015s, `go test ./...` 15.08s, and
    `scripts/check.sh` 1m25.9s (all at `-count=1`, measured at close),
    which the version-bump differential doc will want, the same
    courtesy 2b passed to 2c, 2c to 2d, 2d to 3a, 3a to 3b, and 3b to
    3c.
15. Rung 9 reads four seams, not the plan's two, when it re-verifies
    the dispatcher against its real callers.
16. **The cancellation mechanism beets never specifies, and what this
    pass chose.** A refused offer (the picked waiter departed between
    `pick()` and the offer) tries the next waiter under the same tick.
    Parking the token in the spare slot would strand already-queued
    waiters, who only re-check the spare at registration.

### Non-blocking leftovers, condensed

- `internal/harness/cassette/cassette_test.go:325` names both forms in
  the mutation-1 message; the assertion checks only "not canonical",
  matching an existing sibling test's style rather than tightening it.
- `internal/harness/stub/stub.go:11` and `cassette.go:214`: two doc
  links that could not resolve, fixed in the simplify pass (`7809558`)
  by naming the package in prose instead of a bracketed link.
- `internal/engine/matcher/info.go:229`: `boolField`'s doc comment
  cites a pre-edit line range for `intField`'s comment; worth a
  conductor call since this task exists partly to correct stale
  citations, not yet fixed.
- `internal/engine/musicbrainz/query_test.go:95,111,130`: three
  rendering-table tests carry no row-count guard, so a dropped or
  renamed testdata key would pass vacuously; a `len` check per table
  would close it.
- `internal/engine/musicbrainz/query_test.go:230`: the tenth-gap table
  asserts through `AlbumSearchFor`'s rendered query rather than the
  `[]criterion` slice the plan's table header names; coverage is equal
  or stronger.
- `internal/engine/musicbrainz/client.go:329,376`: `doWithRetry`'s
  signature deviates from the plan's named shape (`requestURL string`
  in place of `*http.Request`), defensible since the body is read and
  closed in one place; no test covers the 404-to-`ErrNotFound`
  mapping.
- `internal/engine/musicbrainz/client.go:404`: `backoff`'s 120s cap is
  unreachable, since the client only sleeps for attempts 1 through 6
  (max 16s); a correct, dead port of urllib3's schedule.
- `internal/harness/conform/request.go:186-220`: `distinctReleaseLookupMBIDs`
  dedups by mbid while `countLookupRequests` counts raw log rows; a
  future case fetching one mbid twice under two `inc` sets would false-
  fire the per-case assertion. No committed case does this today.
- `internal/harness/conform/request.go:105-115`: `summary.Cases`
  increments even on a load or stub-start error, so a broken case still
  counts toward `RequestCaseFloor`; harmless because the gate test
  fatals on any error first, fixed toward the doc comment's promise in
  the simplify pass.

### Filed for other owners

- **The `/ws/2/release/` literal fold**, for the lane owning
  `mbrequest`: `cmd/harness/capture.go:279` and
  `internal/harness/synth/synth.go:41` each spell
  `"/ws/2/release/"` inline rather than reading it from
  `internal/harness/mbrequest`'s own constant. 3b's own guard 6 records
  the deliberate three-copy split between `oracle`, `analysis`, and
  `conform`'s `releaseLookupPrefix`; this is a fourth copy in a
  different package pair, worth folding the day the lane that owns
  `mbrequest` next touches either file.
- **Whether `-race` joins `make check`**, for whoever owns the
  Makefile. This pass's only concurrent code (`limiter.go`) is proven
  race-clean by hand at every task boundary and by the pass-end review,
  but `scripts/check.sh`'s `go test ./...` does not run under `-race`
  today, so a future concurrency bug in `internal/engine/musicbrainz`
  or a later package would not be caught by the gate itself.

### Ruled

**Revision 2.1's five conductor rulings were all taken as recommended.**

1. B1, the reserved token round-robins across the non-empty classes
   below the served one.
2. M-a, `DataSourceMismatchPenalty` is wired into `DefaultTunables`'
   penalty map (Task 3's fix round proved the wiring is real, not a
   coincidence of two equal literals).
3. M-b, `NewLimiter(interval time.Duration)` with `DefaultInterval =
   time.Second`.
4. m22, `corpus/README.md` and `corpus/manifest.json` both carry
   contested-surface rows, confirmed already landed in the parallel
   spec before this pass launched (no-op at this close).
5. m23, the 80%-of-ceiling flag routed to the conductor for the
   2026-09-04 overnight grant's duration only.

**Revision 2.2's ruling** (Task 5's escalation): option (a), the fourth
unexported hook `holdBeforeOffer`, landed as Task 5b, is recorded above
under "What a later pass would be wrong to rediscover" item 15.

The optional `searchlimit` spec clarification at the rung 3 spec `:226`
is **declined**: that line spells legacy keys relative to the plugin
namespace, as its adjacent `source_weight` item does, so
`match.searchlimit` is right as it stands and the ambiguity the
clarification would remove is not worth the edit. 3a's HISTORY
correction #5 (the `SearchLimit` legacy spelling) was already corrected
at `:600` before this pass's close, folded in at the plan's own drafting
(fold m13); no edit was owed here.

### Carried

- **To 3c, now resolved.** `dataSourcePenalty`'s configured-versus-
  literal fallback is **CLOSED** (R15): `defaultDataSourceMismatchPenalty`
  replaces the tunable read and both `conform` compensations are
  removed, with 72/72 tier-0 rows re-verified bit-identical.
  `MergePseudoRelease`'s shallow track-slice copy is **re-owned to 3d
  and rung 9** rather than closed: `AlbumForID` closes the aliasing
  hazard by construction for this pass's own call site, but the
  `slices.Clone` fix is still owed the day a caller retains both the
  merged and the pseudo candidate (item 10 above). The browse fallback
  is a **non-port** this pass, unowned until a corpus case crosses 500
  tracks (unchanged from 3b).
- **To 3d:** the `Preferred.Media`/`Preferred.Countries` config-load
  validation (3d's alone, unchanged from 3b); the compat switch's
  pattern-table selection and the `\b` regime divergence; the tie-
  mirroring work needing both `inclArtist` call shapes; 3c's own R1
  seam (item 3 above), R6's stub blind side (item 4), R7's `Extra`
  producer (item 5), and R11's interim `AdminEmail` contact (item 7c).
- The four-knob wall registration stays the named follow-on task for
  the first rung 3 pass after rung 1b lands (item 2 above); five 3b
  `code-simplifier` leftovers still sit beside it in HISTORY, none
  blocking.
- `profile_hashes` (m1) and the response-status extension, carried
  since 2b and 2c, are unaffected by this pass and stay open.
- Adjudication batch 1 stays fully ruled, 18 `oracle-wrong` and 3
  `oracle-correct` with zero pending; this pass produced no new
  adjudication row.

Budgets: prep 2.55M (two runs, a usage-limit pause), Task 0 0.11M, the
Tasks 1 to 7 chain 1.74M, Task 5b 0.13M, the simplify pass 0.15M, the
final fresh-context review 0.07M, close ≈0.15M (estimate) against the
6M ceiling, not re-derived after Task 4 (the conductor's Task 4
checkpoint found no cause to raise it). Attended interaction points:
planning misses zero; execution sittings one, a combined five-item
grant exchange on 2026-09-05 that covered the unattended overnight
window rather than any 3c content decision.

## Rung 3, pass 3b (The translation layer, the `AlbumInfo` ledger, and the `defect_free` block): `internal/engine/musicbrainz` born, 50/50 ledger rows graded bit-identical (2026-09-04)

Nine plan tasks plus a staleness pre-task, executed sequentially in the
main tree against
`docs/superpowers/plans/2026-09-04-rung-3b-translation.md` (revision
2.1). Commits in order: `946cc21` (Task 1), `a67d5a0` plus `ae2b56e`
(Task 2), `7fcb5b4` plus `7092090` (Task 3), `b8460e0` (Task 4),
`340a983` (the conductor's Task 4 checkpoint), `251ab14` (Task 5),
`cf5599e` plus `40f78be` (Task 6), `8ba42a5` (Task 7), `0621aa3` (the
conductor's plan revision 2.1), `f4eb389` (Task 8), and `096a881` (the
close-out `code-simplifier` pass, 15 files, +115/-157).

What landed. Task 1 built the derived `AlbumInfo` ledger ruling E asks
for: `internal/harness/analysis/album_info.go` with the nineteen named
subjects, the `AlbumInfoLedgerRowFloor` of 50, the six-part
`CheckAlbumInfoLedger`, and the pointer-per-scalar ledger types;
`derive_album_info.py` in `derive_ties.py`'s shape with its
oracle-tagged Go driver; and `corpus/analysis/album_info.json` at 19
subjects, 50 rows, 4 merged blocks, 262,377 bytes. Task 2 landed the
`defect_free` block on `AdjudicationEntry` with `CheckDefectFree`, the
three within-row rules, `matcher.SDPatternsNoBracketDiscount` as the
third pattern table, `derive_defect_free.py` with its module-object
monkeypatch, and the first population of the ten string-heuristic rows.
Task 3 birthed `internal/engine/musicbrainz` with `doc.go`'s charter,
the WS/2 release vocabulary, `wireInt`'s two-shape position decode,
`Validate`'s boundary rules, and `GroupRelations`; it also added
`matcher.Tunables.VAName`. Task 4 added `preferredAlias`,
`keyWithPreferredAlias`, `parseArtistCredits`, and
`applyVariousArtists`. Task 5 added `tracksFromMedium`, `buildTrack`,
`extractTracks`, and `deriveMedia` with both `"Media"` sentinel routes.
Task 6 added `preferredReleaseEvent`, `parseDate`, and `TranslateAlbum`,
the one album-side composition. Task 7 added `PseudoReleaseTarget` and
`MergePseudoRelease`. Task 8 landed the translation tier in
`internal/harness/conform`: `CheckTranslatedAlbum`'s field-by-field
comparison over `errors.Join`, `gradeLedger` with its compared-row
count, the twenty-one-row nullability policy, the profile-name overlay
with its key-drift guard, the browse-fallback guard, and the clock walk
widened from `internal/engine/matcher` to all of `internal/engine`.

What the gate caught: `bash
scripts/check.sh` stayed green at every task end and found nothing
novel. That is 2c's, 2d's, and 3a's record repeating. Three tasks took
a fix round and one escalated, all four from the in-chain
`diff-reviewer` rather than from the gate. Task 2's reviewer blocked on
an unimplemented acceptance clause: `derive_defect_free.py` solved the
optimal item-to-track mapping under the substituted table and then
discarded it, so the plan's "verify rather than assume the mapping is
unchanged" went unproven; `ae2b56e` emits the mapping as a row field and
asserts it against each case's committed `Expected.Assignment`, and all
ten resolve to `{0:0, 1:1, 2:2}`. Task 3's reviewer blocked because the
four `Validate` failure tests asserted only `errors.Is` and never read
the message, so stripping the index and id from every format string
would have left the suite green against a doc comment promising to name
the offending element; `7092090` added the four substring assertions and
proved them by stripping the identifiers. Task 6's reviewer blocked
because `TranslateAlbum`'s doc comment dropped the plan's error-contract
paragraph, leaving an exported function that returns an error with no
godoc about when it does; `40f78be` restored it. Task 5's reviewer
escalated rather than blocking: the plan's named mutation 4, dropping
the `preferredAlias` guard on the track-title override, is provably
unobservable through `TrackInfo.Title`, because `keyWithPreferredAlias`
resolves the same aliases under the same tunables immediately
afterward. The implementer declared the non-firing instead of
fabricating a proof. The reviewer reproduced it independently. The
conductor ruled the guard stays, its non-effect is stated in `tracks.go`,
and the mutation is struck from the proof set as a declared unobservable
mutation rather than a coverage gap (plan revision 2.1, `0621aa3`); Task
5 was accepted at `251ab14` with no code change.

### Translation coverage table

Sixteen contracts, T1 through T16, each with its Go site and what grades
it. A ledger row grades a contract when one of `album_info.json`'s 50
rows exercises it through `TranslateAlbum`; the rules the corpus cannot
reach are graded by hand-authored fixtures in the owning package's own
`testdata`, per the plan's fixture-proximity rule.

| Contract | Go site | Graded by | Corpus reach |
| --- | --- | --- | --- |
| T1 one album-side function, merge inside | `translate.go` `TranslateAlbum`, `pseudo.go` `MergePseudoRelease` | all 50 ledger rows; the 4 `Merged` blocks | Yes |
| T2 key normalization, one rule surviving | `release.go` struct tags; `alias.go` `keyWithPreferredAlias` | all 50 ledger rows; `alias_test.go`'s table | Yes |
| T3 `mediums` counts every medium | `translate.go` (`len(rel.Media)`), `tracks.go` `extractTracks` | ledger rows; `translate_test.go`'s two-medium release under `IgnoredMedia` | Yes |
| T4 the `"Media"` sentinel, two routes | `tracks.go` `deriveMedia` | 4 zero-track rows (`real-nonlatin-hamasaki-ayumi`), 2 mixed-format rows (`real-noise-dylan-bootleg`); 4 sentinel fixtures | Yes, both routes |
| T5 assembly order and the two filters | `tracks.go` `tracksFromMedium` | `gap_pregap_medium.json`, `gap_data_tracks_medium.json`, `gap_bracket_title_in_tracks.json` | No: pregap and data tracks are declared gaps 1 and 2 |
| T6 the three track-level overrides | `tracks.go` `buildTrack` | all 50 ledger rows' track walks; the truthiness-fallback tests | Yes, except the title guard (declared unobservable) |
| T7 `_preferred_alias` | `alias.go` `preferredAlias` | `preferred_aliases.json`, T7's eight rows | Thin: declared gap 5, one profile and one case |
| T8 credit flattening | `credit.go` `parseArtistCredits` | `artist_credit_alias.json`; ledger rows | Yes |
| T9 the VA override | `credit.go` `applyVariousArtists` | `translate_test.go`'s VA release; ledger rows | Yes |
| T10 release-event selection | `events.go` `preferredReleaseEvent` | 4 preferred-country fixtures | No: declared gap 3, the fallback only |
| T11 date parsing | `events.go` `parseDate` | T11's six-row table, `"1999-XX-06"` proving no early return | Yes, through the fallback path |
| T12 the pseudo merge, seven fields moved | `pseudo.go` `MergePseudoRelease`, `PseudoReleaseTarget` | the 4 `Merged` blocks; 5 `pseudo_target_*.json` fixtures | Partly: declared gap 4 on year and mediums |
| T13 no Unicode normalization | `doc.go` (a stated non-check) | nothing to grade; the absence is the contract | n/a |
| T14 config reads inside the layer | `matcher/tunables.go` (`VAName` and the rest) | `TestDefaultTunables_VAName`; the defaults cross-check | Yes |
| T15 the plugin's cached config reads | `derive_album_info.py`, `derive_defect_free.py` (one plugin built after config, never mutated) | the derivation's own shape asserts | n/a, derivation-side |
| T16 the browse fallback, not ported | `doc.go`'s refusal; `conform`'s `checkNoReleaseReachesBrowseThreshold` | `TestNoCommittedReleaseReachesTheBrowseThreshold` over all 215 cases | Yes, as a guard |

**Six declared corpus-coverage gaps, recorded as gaps rather than as
passing checks** (spec, M8).

1. **Pregap**: zero committed cases carry a non-null `pregap`.
2. **Data tracks**: zero committed cases carry a non-empty
   `data-tracks` array, so both the `IgnoreDataTracks` branch and the
   `"[data track]"` title filter are ungraded by the corpus.
3. **The preferred-country release-event branch**: the branch is
   reachable data-side and unreached only for that reason. Of the 303
   release-lookup responses, 290 of 298 top-level release events carry
   a non-null `area` with `iso-3166-1-codes`, across 208 case
   directories. A fourth profile setting `match.preferred.countries`
   would reach the branch immediately, and `various-artists` is the
   concrete target.
4. **The merge's `year`, `original_year`, and `mediums` path**: equal
   on both sides of all four committed pseudo cases, so the corpus
   gives that path no signal.
5. **The alias path's breadth**: `import.languages` is non-empty in
   exactly one committed profile and one committed case, so 214 of 215
   cases run with `preferredAlias` returning nil unconditionally.
6. **Two boundary rules with no corpus instance**: no committed
   release carries an empty resolved track credit and no committed
   backward `transl-tracklisting` relation lacks its `release` target,
   so both of `Validate`'s rules are graded by hand-authored fixtures
   alone. Each closes a Go panic rather than an error. That is why the
   absence is worth recording.

Together with 3a's three
(`match.required`, `match.ignored`, `preferred.original_year`), **rung 3
now carries nine declared gaps**; 3d's close should read that total
rather than re-count it.

**Task 5's struck mutation is not a gap.** The `preferredAlias` guard on
the track-title override is provably redundant with respect to any
observable `Title`, so no fixture can make its removal fire. The guard
is faithful to beets, it stays, and `tracks.go` documents its
non-effect at the site.

### Evidence

The `AlbumInfo` ledger: 19 subjects, exactly 50 rows, 4 merged blocks,
262,377 bytes, `shim_version` `"1"`, `beets_version` `2.13.1`. Its
derivation ran 25.19s for the re-derive test and 25.26s for the
perturbation test at `-count=1`. Row ids are the deduplicated
`/ws/2/release/<mbid>` lookups of each subject's own case, and no
subject case logs a duplicate lookup.

The translation tier, `TestTranslationLedgerGradesCleanly` against the
real `corpus/analysis/album_info.json`: **50 of 50 rows graded, zero
disagreements, on the first pass**, in 364 to 390ms across runs
(≈7.3 to 7.8ms per row). The compared-row count equals the ledger's row
count, so a silently skipped row fails. 561 of the 650 album-level
scalars in the ledger are non-null, so the grading is not vacuous. All four
`Merged` blocks graded `MergePseudoRelease` as well as the translation.

The `defect_free` block, ten rows, derived float64 beside the prose
figure each row's `corrected_verdict` carries:

| Row | Derived distance | Prose figure | Table |
| --- | --- | --- | --- |
| `adj-adv-ft-draft-colon` | 0.10185185185185186 | 0.101852 | `SDPatternsWordBoundary` |
| `adj-adv-ft-left-colon` | 0.11764705882352941 | 0.117647 | `SDPatternsWordBoundary` |
| `adj-adv-ft-gift-space` | 0.10784313725490197 | 0.107843 | `SDPatternsWordBoundary` |
| `adj-adv-ft-soft-onesided` | 0.08666666666666667 | 0.086667 | `SDPatternsWordBoundary` |
| `adj-adv-ft-craft-titles` | 0.09170751633986929 | 0.091708 | `SDPatternsWordBoundary` |
| `adj-adv-ft-theft-artist` | 0.0909090909090909 | 0.090909 | `SDPatternsWordBoundary` |
| `adj-adv-remix-both-sides` | 0.08 | 0.080000 | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-onesided` | 0.07142857142857142 | 0.071429 | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-bracket` | 0.0925925925925926 | 0.092593 | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-track-titles` | 0.08187134502923976 | 0.081871 | `SDPatternsNoBracketDiscount` |

The both-ways check ran the same ten subjects twice in one process, once
with the pattern table untouched and once with it substituted. The
baseline leg reproduced every case's committed `case.json` distance
bit-identically. That is what makes the substituted leg's figures
trustworthy. All ten re-solved mappings are `{0:0, 1:1, 2:2}`, matching
each case's committed `Expected.Assignment`, so the substitution moves
distances and never the assignment. The write-back added exactly ten
`defect_free` blocks and changed no pre-existing field on any row.

`make derive` and the byte-identity invariant: the oracle-tagged suite
ran twice in full, at 983.4s under Task 1 and at 17m50.8s under Task 2,
both exit 0, re-deriving `components.json`, `ties.json`,
`file_tags.json`, `album_info.json`, and every `case.json`
byte-identically. The `code-simplifier` reviewer re-ran the
`AlbumInfo|DefectFree` subset at 64.9s with `corpus/` unchanged
afterward.

The browse-fallback guard scans all 215 cases in 2.15 to 2.38s as
reported and ≈3.4s per scan on the reviewer's machine, where it
dominated the `conform` package's 7.7s. It runs twice. Nothing reaches
the threshold: the two `real-boxset-rachmaninov-complete` twins are the
corpus maximum at 470 tracks against `BROWSE_MAXTRACKS` of 500, and
lowering the test-local threshold to 400 fires naming both.

Gate and suite wall times, all at `-count=1`. `bash scripts/check.sh`
ran green at every task end, 57.7s to 1m2.3s across tasks, with gitleaks
self-reporting 52.9s to 1m8.8s over ~113.8 MB and dominating as it did
in 3a. `go test ./...` ran 8.563s and 8.640s on two tasks; the
`internal/...` subset ran 16.6s on a third. Tier-0 replay stayed green
over all 72 rows throughout, since it runs inside `go test ./...`.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, verbatim from the plan's Spec section,
   travels forward from every rung 3 pass close: "Rung 15 re-runs
   `internal/harness/conform`'s three tiers over the complete corpus, at
   the matcher commit as landed, and its acceptance is the launch bar's
   matcher clause read against that run: zero false confident matches
   over every stratum including stock-config, and zero unadjudicated
   beets-right/Go-wrong cases. Rung 3's own acceptance is provisional
   against the corpus at its then-current size. Any tier-2 divergence
   carrying a pending row at rung 15 is the launch gate's business, not
   rung 15's to wave through."
2. **3a's R1 carry is closed, and this is its resolution.** An absent
   source field becomes the plain-string zero value, never a nil-bearing
   vocabulary. The proof is source-scoped: `internal/engine/musicbrainz`
   is the only translation source rung 3 ships, and every field it
   writes comes from a Go decode that yields `""` rather than a null. The
   thing that would reopen the question is a second metadata source, and
   the translation tier's three `nullStrict` fields are what would catch
   it, because a ledger null against a Go zero fails there rather than
   passing silently.
3. **The pointer-per-scalar ledger convention.** `LedgerAlbumInfo`'s and
   `LedgerTrackInfo`'s scalars are `*T` because a JSON null and an empty
   string are different facts about what beets produced, and the
   `null-and-empty` testdata fixture exists to prove the two decode
   differently. The convention is deliberate against the
   go-conventions no-pointer default and its reason is on the type.
4. **The nullability policy is twenty-one rows, three regimes.**
   `conform`'s `fieldNullability` splits the fourteen `AlbumInfo` fields
   (less `Tracks`) plus the eight `TrackInfo` fields into 3 `nullStrict`,
   9 `nullNeverExpected`, and 9 `nullEqualsZero`, one reason per row, and
   `TestNullabilityPolicyCoversEveryField` asserts the count against
   `reflect.TypeFor` so a new vocabulary field cannot slip past. The
   table documents; the comparison passes its policy as a literal at each
   call site. Nothing catches a drift between the two. The limit is
   known and accepted.
5. **`Track.Position` is a two-shape decode and has to be.** The
   MusicBrainz payload spells `position` two ways in this corpus, 6,763
   integers and 297 strings, and it is the only field that does.
   `wireInt` is the narrow answer, proven directly against
   `corpus/cases/media-baseline/responses/001.json`, whose six positions
   are all JSON strings. 3c's live client meets real captures only, where
   the string form does not occur, so the type is carried for the corpus
   rather than for the wire.
6. **The profile overlay is a name-to-tunables table, not a YAML
   parse.** No YAML parser exists in the module and the plan forbids
   adding one, so `conform`'s `mergeProfile` overlays by profile name
   (`stock` and `estate` no-ops, `locale` setting `Languages` to
   `["ja"]`) and fails on any other name. A line-oriented drift guard
   scans `corpus/profiles/*.yaml` against a ten-key allowlist, five keys
   with a `Tunables` home and five without, so a profile that grows a key
   the overlay does not model reds the gate instead of being silently
   ignored.
7. **The monkeypatch target is the module object.**
   `beets.autotag.distance` resolves to a function, so
   `derive_defect_free.py` reaches `SD_PATTERNS` through
   `importlib.import_module` on the module object and restores it in a
   `finally`. `string_dist` is uncached and reads its module global at
   call time. That is why one process can carry all ten subjects with a
   per-subject substitution and no leak. Editing `derive_oracle.py`,
   `derive_component.py`, or the venv would break the byte-identity
   invariant. Do not do it.
8. **The merge moves exactly seven fields.** `MergePseudoRelease` copies
   the pseudo by value and overwrites `Media`, `Mediums`, `Country`,
   `CatalogNum`, `Year`, `OriginalYear`, and `Label` from the actual;
   `Artist`, `Album`, `VA`, `AlbumDisambig`, `AlbumID`, `DataSource`, and
   `Tracks` stay the pseudo's. All fourteen are asserted. The copy is
   shallow, so `merged.Tracks` aliases the pseudo's backing array, and a
   3c caller that mutates a merged track element writes through to the
   pseudo candidate.
9. **The `Preferred.Countries` two-regime hazard.** One config list feeds
   two different rules. `internal/engine/musicbrainz` does plain
   case-sensitive membership; `internal/engine/matcher` compiles an
   anchored case-insensitive regex. Both sites name the divergence, and
   a mutation compiling the membership test case-insensitively fires.
10. **The three `SD_PATTERNS` tables and what each is for.**
    `BeetsSDPatterns` is shipped and defective on purpose;
    `SDPatternsWordBoundary` carries the six `adv-ft-*` figures;
    `SDPatternsNoBracketDiscount` carries the four `adv-remix-*` figures
    and is born here as `BeetsSDPatterns` less its two 0.3-weight bracket
    rows. 3d's compat switch selects one by assignment to
    `Tunables.StringDistPatterns` and edits no component. 3a's warning
    travels with them: the word-boundary guard changes zero of the 72
    tier-0 rows, so tier 0 is blind to a "fixed" table.
11. **The `\b` regime divergence.** `SDPatternsWordBoundary`'s `\b` is
    Unicode-aware in Python's `re` and ASCII-only in Go's RE2, so 3d's
    compat switch and the `defect_free` derivation can differ on
    non-ASCII input. All six `adv-ft-*` pairs are ASCII, so nothing moves
    today. It travels beside 3a's `SD_PATTERNS`-as-a-value seam.
12. **The `defect_free` block as built.** It carries `distance`,
    `recommendation`, `action`, and `pattern_table`; the derivation's own
    unrounded float64 is the stored value and the prose figure is the
    cross-check. `StringHeuristicHypotheses` is the predicate selecting
    which adjudication rows must carry one, and the six
    `adv-assign-*` rows are excluded because their hypothesis is not a
    string-heuristic one. The two legal `pattern_table` values are the
    named tables above. `CheckDefectFree` enforces the categorical
    cross-check both ways, including the forbidden direction: a block on
    a row whose hypothesis is not a string-heuristic one is a failure.
    The numeric check was one-time, at population.
13. **The `AlbumInfo` ledger as built.** It is keyed on
    `(case_id, release_mbid)`, carries nineteen subjects and 50 rows, and
    grades the translated fields only. Applied fields are deliberately
    out of scope, so **rung 8 must not read it as an applied-field
    ledger**; `file_tags.json` is that. Its `Merged` block appears on
    exactly the four `AlbumInfoPseudoSubjects` cases and check 6 reads
    that named list rather than deriving its own expectation. 3d should
    not re-key it.
14. **The browse fallback's refusal and its guard.** `doc.go` states that
    `_ensure_complete_recordings` is not ported;
    `BROWSE_MAXTRACKS` is 500 and the corpus maximum is 470, a 30-track
    headroom on the two rachmaninov twins. `conform`'s guard uses `>=`
    rather than beets' `>`, deliberately, because its purpose is to catch
    a corpus case approaching the line. **3c owns the port** if a case
    ever crosses it.
15. **`Tunables.VAName` landed and `AliasesAsCredits` did not.** The
    exclusion is deliberate per R8 and is stated on `VAName`'s own doc
    comment, so a later pass does not add the second knob for symmetry.
16. **The wire-type census, restated for 3c.** `position` is the only
    two-shape field; see guard 5. Nothing else in the WS/2 vocabulary
    this pass models needs a custom unmarshaller.
17. **Measured wall times for the version-bump differential doc**, the
    courtesy 2b passed to 2c, 2c to 2d, and 2d to 3a: the `AlbumInfo`
    derivation 25.2s, the full oracle suite 983.4s and 17m50.8s on two
    runs, the translation tier 0.37 to 0.39s for 50 rows, the browse scan
    2.2 to 3.4s per pass over 215 cases, `go test ./...` 8.6s, and
    `scripts/check.sh` 57.7s to 1m2.3s with gitleaks the dominant leg.

### Carried

- **To 3c:** `dataSourcePenalty` falls back to
  `Tunables.DataSourceMismatchPenalty` where beets falls back to the
  class literal `0.5`; identical under shipped config, diverging only if
  `source_weight` is set (3a's carry, unchanged). New from this pass:
  `MergePseudoRelease`'s shallow copy aliases the pseudo's track slice,
  so a 3c caller that mutates a merged track writes through; and the
  browse fallback stays unported until a corpus case crosses 500 tracks.
- **To 3d:** the `Preferred.Media` and `Preferred.Countries` compile
  panics want a wrapped error at config-load time. **3a addressed this to
  3b or 3d; it is now 3d's alone**, because `internal/config` has no
  validation hook and 3b compiles no regex, so 3b was never its home.
  Also 3d's: the four-knob wall registration
  (`StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, `SearchLimit`),
  which stays the named follow-on task for the first rung 3 pass that
  runs after rung 1b closes, with `internal/config/keys.go` the file no
  lane edits until then; the compat switch's table selection and the
  `\b` regime divergence (guards 10 and 11); and the tie-mirroring work
  that needs both `inclArtist` call shapes (3a's guard 7).
- 3a's first carried item is closed by this pass. R1's absent-field
  resolution is guard 2 above, and no divergence is owed.
- Five `code-simplifier` leftovers, each seen and left deliberately:
  `wireInt`'s two decode errors do not wrap `ErrMalformedRelease`, while
  that sentinel's comment calls it the one sentinel every boundary
  failure wraps; `releaseLookupPrefix` now exists in three copies
  (`oracle`, `analysis`, `conform`), deliberately unshared so three
  independent walks never share a decision, which is worth revisiting if
  a fourth appears; `derive_album_info.py`'s `ALBUM_FIELDS` comment says
  "fourteen" over a thirteen-entry tuple; `Validate`'s data-track walk is
  observed by no test, a pre-existing gap that flipping its
  `withDataTracks` argument leaves green; and the m11 pattern-table guard
  now checks only the ledger-to-matcher direction.
- `profile_hashes` (m1) and the response-status extension, carried since
  2b and 2c, are unaffected by this pass and stay open.
- Adjudication batch 1 stays fully ruled, 18 `oracle-wrong` and 3
  `oracle-correct` with zero pending, and this pass produced no new
  adjudication row. **Four `corpus/README.md` cells that still described
  the batch as pending are corrected in this close**, along with the
  `stock.yaml`/`locale.yaml` placeholder claim and the
  regeneration-command list, which gains `album_info.json`'s command.

Budgets: ≈4.82M subagent tokens (the plan-prep workflow 2,106K across
three inventory readers, synthesis, a draft, two reviews, and the fold;
Task 0 109K; the Tasks 1 to 8 chain 2,057K across 22 agents;
`code-simplifier` 164K and its review 61K; the close draft 171K, its
prose review 62K, and the fold 90K) against the 6M ceiling. The 4.8M
flag was crossed by the fold, the pass's final dispatch, after every
deliverable had landed; the plan-prep workflow is the 2.1M the 3a
Budgets line never counted, and without it execution closed at 2.7M. Attended interaction points: zero. Planning
questions zero, since the plan was approved under the standing grants;
execution sittings zero; planning misses zero. Task 5's struck mutation
was a plan defect. Execution catches that shape of defect and no owner
question would have reached it, because the defect was invisible
without the code in front of it. The four conductor rulings taken at plan
landing (the nineteenth ledger subject, the three contested-surface rows,
the spec's tree-column fix, and STATUS's 3a spend figure) cost no
attended time.

## Rung 3, pass 3a (The scoring core, `string_dist`, and tier 0): 21 distance components, `internal/harness/conform` born (2026-09-04)

Nine plan tasks executed sequentially against
`docs/superpowers/plans/2026-09-04-rung-3a-scoring-core.md` (revision
2.1). Commits in order: `28396a7` (Task 1), `73d5321` (Task 2), a
same-session model restart pause at `7a21dcf`, `b5de09e` (Task 3),
`0592a32` (Task 4), `46c2eb6` (Task 5), `c5230ae` (Task 6), `f5b3243`
plus `ca50ff6` (Task 7), `91a16a4` plus `57bc094` (Task 8), `8d5a793`
(Task 6b, scope added mid-pass), and `05575d2` (the close-out
`code-simplifier` pass, 18 files, +205/-223).

What landed: `internal/engine/matcher`, a new package holding
`Tunables` (pinned beets 2.13.1's shipped `match:` defaults with a
citation per field) and the `AlbumInfo`/`TrackInfo`/`ItemFields`
vocabulary (Task 1); the `Distance` accumulator with seven of the
eight adders, both penalty views (`AddedKeys`/`PerKey` versus
`NonzeroItems`/`NonzeroKeys`), and the insertion-ordered summation
contract 5 requires (Task 2); the embedded `unidecode` 1.4.0
transliteration table via a committed one-time generator,
`tools/unidecode-table/generate.py` (Task 3); `StringDist` and the
eighth adder, `AddString`, plus the `SD_PATTERNS` pattern loop and its
byte-identity guard, which tier 0 cannot supply (Task 4); the seven
track-level components and `DistanceForTrack` (Task 5);
`internal/harness/conform`, a new package this pass creates,
with the typed `components.json` decoder (tier 0), the M1 defaults
cross-check, the `unidecode` manifest pin check, and the
engine-never-imports-harness / no-`time.Now` architecture checks
(Task 6); the fourteen album-level components (thirteen files,
`album_extras.go` the named one-file-two-component exception),
`MostCommonTags`, and `score.go`'s `albumComponents` composition
(Tasks 7 and 8); five reviewer-hardening follow-ups to the decoder
(malformed-row and mistyped-field failures, an add-shape coverage
test, a thresholds mutation test, and testdata-subtree exclusions for
both arch-check walks), landed as Task 6b, scope the conductor added
mid-pass from the Task 6 reviewer's non-blocking list rather than
carried to 3b (Task 6b).

AddString landed as a revision-2.1 editorial fold. The plan's Task 4
deliverable count moved 5 -> 6; no new work entered the pass.
Task 2's `diff-reviewer` escalated that the eighth adder was deferred
to Task 4 (`strdist.go`, alongside `StringDist` itself) but the plan's
Task 4 section never named it; the conductor accepted the escalation
and ruled the deferral correct (a placeholder would have been the
assertion-discipline rule's forbidden silent-success fake), then added
the two-line revision note at the plan's top, landed in this pass close
alongside the ledgers.

What the gate caught: nothing novel in `bash scripts/check.sh` itself
across all nine tasks; the pass's own fixtures are what proved a
mutation fires, the same discipline 2c and 2d both record. Two
findings worth a future pass's attention surfaced from hand-authored
mutation testing rather than the gate: `component_test.go`'s
`aggregate/sorted-penalties-tied` fixture cannot discriminate a
`(distance, key)`-ascending sort from a dropped sort at all (both an
insertion-ordered album and a key-ascending sort tie on that row's own
data), so the mutation the plan's M-b fold specified does not fire;
the implementer substituted a full key-order reversal
(`cmp.Compare(b.Key, a.Key)` in place of `cmp.Compare(a.Key, b.Key)`),
which does discriminate, and the `diff-reviewer` confirmed the
substitution independently rather than taking the report's word for
it. Task 4's Go/Python case-mapping sweep, run over every one of
the 1,112,064 valid Unicode code points against pinned beets in
`tools/venv/.venv`, found 56 where Go's simple-case `strings.ToLower`
disagrees with Python's full-case `str.lower`; 54 agree again once the
pipeline's post-transliteration `[a-z0-9]` strip runs, and two,
U+A7CB and U+A7DC, do not: both are unassigned in CPython 3.13's
Unicode 15.1 but lowercase under Go 1.27's Unicode 17.0, and
`unidecode` 1.4.0 maps their lowered forms to `"u"`/`"l"` where the
unlowered originals strip to nothing. This is a toolchain-Unicode-
version dependence the study never claimed and the port now carries
explicitly (`internal/engine/matcher/testdata/case-mapping-sweep.tsv`
names both rows so a third divergence reds the gate rather than
passing silently).

### Component coverage table

Twenty-one components: seven track-level (`internal/engine/matcher/track.go`'s
fixed order) and fourteen album-level (`score.go`'s fixed order).
`components.json`'s eleven tier-0 methods carry no album-distance rows
at all and no `track_distance` row that sets a length, index, MB track
ID, or disc number, so "tier-0 row?" is `No` for eighteen of the
twenty-one; every `No` still carries its own fire-proving and
decline-proving unit test in the component's own file, per contract 8
and the plan's assertion-discipline rule, distinct from the three
M8-declared corpus-coverage gaps below.

| Component | File | Weight key | Fire condition | Tier-0 row? |
| --- | --- | --- | --- | --- |
| track_length | `track_length.go` | `track_length` | `info.Length != 0`; `AddRatio` on `abs(item.Length-info.Length)-Grace` over `Max` | No (own fire/decline test) |
| track_title | `track_title.go` | `track_title` | always | Yes (`track_distance/*`, e.g. `identical`) |
| track_artist | `track_artist.go` | `track_artist` | `inclArtist && info.Artist != "" &&` item artist not in `VA_ARTISTS` (case-insensitive) | Yes (`track_distance/different-artist`, `tolerate-va`, `identical`) |
| track_index | `track_index.go` | `track_index` | `info.Index != 0 && item.Track != 0`; tolerates per-disc or absolute numbering | No (own fire/decline test) |
| track_id | `track_id.go` | `track_id` | `item.MBTrackID != ""` | No (own fire/decline test) |
| medium (track) | `track_medium.go` | `medium` | `info.Medium != 0 && item.Disc != 0` | No (own fire/decline test) |
| data_source (track) | `track_data_source.go` | `data_source` | always calls `AddDataSource` (its own three-way condition, contract 6) | Yes (`track_distance/data_source/*`, 10 rows) |
| artist | `album_artist.go` | `artist` | `!info.VA` | No (own fire/decline test) |
| album | `album_album.go` | `album` | always | No (own fire/decline test) |
| media | `album_media.go` | `media` | `info.Media != ""`; priority branch when `Preferred.Media` set, else equality against `likelies.Media` | No (own fire/decline test; anchor proved by `ca50ff6`'s dedicated rows) |
| mediums | `album_mediums.go` | `mediums` | `likelies.DiscTotal != 0 && info.Mediums != 0` | No (own fire/decline test) |
| year | `album_year.go` | `year` | three branches in order (preferred-original, likely-vs-release, decline); see below | No (own fire/decline test; branch 1 is a declared gap) |
| country | `album_country.go` | `country` | priority branch when `Preferred.Countries` set and `info.Country != ""`, else equality when both `likelies.Country` and `info.Country` set | No (own fire/decline test; anchor proved by `ca50ff6`'s dedicated rows) |
| label | `album_label.go` | `label` | `likelies.Label != "" && info.Label != ""` | No (own fire/decline test) |
| catalognum | `album_catalognum.go` | `catalognum` | `likelies.CatalogNum != "" && info.CatalogNum != ""` | No (own fire/decline test) |
| albumdisambig | `album_disambig.go` | `albumdisambig` | `likelies.AlbumDisambig != "" && info.AlbumDisambig != ""` | No (own fire/decline test) |
| album_id | `album_id.go` | `album_id` | `likelies.MBAlbumID != ""` | No (own fire/decline test) |
| tracks | `album_tracks.go` | `tracks` | always runs the pairing loop (zero pairs adds nothing) | No (own whole-distance tests, `91a16a4`) |
| missing_tracks | `album_extras.go` | `missing_tracks` | `len(info.Tracks) > len(pairs)`, one penalty per excess candidate track | No (own fire/decline test) |
| unmatched_tracks | `album_extras.go` | `unmatched_tracks` | `len(items) > len(pairs)`, one penalty per excess item | No (own fire/decline test) |
| data_source (album) | `album_data_source.go` | `data_source` | always calls `AddDataSource` | No (own fire/decline test) |

**The three corpus-unreachable surfaces are a declared corpus-coverage
gap (spec, M8). No fixture covers them, and none is owed in 3a.**
`match.required` and `match.ignored` carry their defaults and
citations on `Tunables` (`tunables.go:80-92`). The admission filters
that read them are 3d's, and 3a lands no consuming component for
either. Their coverage is 3d's table-driven unit tests, named in
`Tunables`' own doc comment. `preferred.original_year` is 3a's own:
it is `addAlbumYear`'s branch 1 (`album_year.go`), graded by a
hand-authored table
(`internal/engine/matcher/testdata/album_year_preferred_original.json`,
three rows) because no committed corpus case sets
`preferred.original_year`. The mutation that proves it fires:
`TestAddAlbumYearPreferredOriginalYearTakesPrecedence` asserts branch
1 runs, not branch 2's zero, when `Preferred.OriginalYear` is set even
though branch 2 would also apply; `f5b3243`'s own commit message
separately names the table's own proof, "a short-circuit mutation
that fails all three of its rows" (a mutant that returns after
checking only `info.Year != 0` reds the whole fixture).

### Evidence

Tier-0 replay (`go test -count=1 -v ./internal/harness/conform/...
-run TestComponentTableReplaysAllRows`), measured now: 72 of 72
committed rows replayed, all bit-identical, zero near-misses at the
1e-12 fallback. Per method: `add` 8, `add_data_source` 4,
`add_equality` 3, `add_expr` 2, `add_number` 4, `add_priority` 4,
`add_ratio` 4, `add_string` 3, `generic_penalty_keys` 1, `string_dist`
23, `track_distance` 16. Total 72, matching `ComponentRowFloor`
exactly. Defaults cross-check
(`TestCommittedCasesEchoTunablesDefaults`), measured now:
weights and thresholds cross-checked against all 215 committed cases,
zero divergences (no committed profile sets a `match:` key). Full
suite (`go test -count=1 ./...`): 12 packages, all green, real 13.4s.
`bash scripts/check.sh`, measured now over two consecutive runs: total
wall time 1m45.6s (gitleaks self-reporting 1m39.6s / 99.6s) and, on a
second run minutes later, 3m24.6s (gitleaks self-reporting 3m13s).
Gitleaks' full-text scan of the tracked worktree (`--no-git`,
~112.9 MB, `corpus/cases/` the dominant contributor) is both the
gate's dominant cost by a wide margin and its most variable one,
roughly doubling between two runs with no tree change in between. The
version-bump differential doc's next update should budget against the
higher figure rather than the first sample. The remaining ~6-12s
across both runs covers `make check` (go vet, golangci-lint 0 issues,
`go test ./...` cached from the fresh 13.4s run above, `go mod tidy`
clean), shellcheck, ruff (6 tracked `*.py` files, all formatted, all
passing), vale over Go comment prose, `docker compose config`, and the
pattern-filtered `systemd-analyze verify`, combined; none of those
self-report an individual wall time in check.sh's own output. **No
`make derive` ran anywhere in this pass**: no task touched oracle
territory, Task 1's venv read re-verified the M2 resolution and
Task 3's venv read drove the one-time `unidecode` generator, and
neither is a derivation.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, verbatim from the plan's Spec
   section, travels forward from every rung 3 pass close: "Rung 15
   re-runs `internal/harness/conform`'s three tiers over the complete
   corpus, at the matcher commit as landed, and its acceptance is the
   launch bar's matcher clause read against that run: zero false
   confident matches over every stratum including stock-config, and
   zero unadjudicated beets-right/Go-wrong cases. Rung 3's own
   acceptance is provisional against the corpus at its
   then-current size. Any tier-2 divergence carrying a pending row at
   rung 15 is the launch gate's business, not rung 15's to wave
   through."
2. **The four-knob wall registration is a named follow-on task** for
   the first rung 3 pass that runs after rung 1b closes:
   `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, and
   `SearchLimit` are named on `Tunables` today with citations but
   registered on no operator wall, because `internal/config/keys.go`
   still carries `ownedWall` as a slice literal. No lane may edit that
   file until rung 1b lands.
3. **The `SD_PATTERNS`-as-a-value seam**: `BeetsSDPatterns`,
   `SDPatternsWordBoundary`, and `Tunables.StringDistPatterns`
   (defaulting to `BeetsSDPatterns`) let 3b's `defect_free` derivation
   and 3d's compat switch swap a table rather than edit a component.
   Both measured figures travel with the warning that produced them:
   the word-boundary guard changes **zero of the 72 tier-0 rows**, so
   tier 0 is blind to a "fixed" table. Task 4's byte-identity test
   against beets' six shipped literals is the only check that catches
   a silent improvement to the table.
   `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")` reads
   `0.061111111111111116` under `BeetsSDPatterns` and
   `0.6111111111111112` under `SDPatternsWordBoundary`.
4. **The M2 resolution and its two drafted `corpus/README.md` cells**
   are landed in this close: `original_date` and `per_disc_numbering`
   change from "verdict-moving: ..." to "applied-field-moving, never
   verdict-moving: ..." in the config-profile transform table
   (`corpus/README.md`), and the stock-config stratum finding's own
   single-accessor sentence is corrected to name both accessors
   (`AlbumInfo.item_data` for `original_date`, `TrackInfo.raw_data`
   for `per_disc_numbering`). Also landed: the mechanics study's Q4
   "19 keys" claim is superseded. The tree and
   `config_default.yaml:174-193` (twenty entries; `:173` is the
   header) agree on twenty `distance_weights` keys, matching all 215
   committed cases' own echoed weight sets.
5. **Four corrected citations**, so no later pass re-derives them:
   `DataSourceMismatchPenalty`'s default is `metadata_plugins.py:180`
   wired at `:202`, not `:190-195`; the `SearchLimit` legacy spelling
   is `musicbrainz.match.searchlimit` (corrected at 3c's plan fold), not
   a top-level `match:` key;
   `ExtraTags` lives on the musicbrainz plugin with no
   `config_default.yaml` citation; and `match:` carries
   `album_disambig_fields` / `singleton_disambig_fields`, display-only
   and deliberately outside `Tunables`.
6. The decoder contract as built, for 3c and 3d to extend rather
   than re-derive from `derive_component.py` a second and third time:
   `internal/harness/conform`'s decoder dispatches on `method`, not
   row id, across eleven methods; method `add` splits on an
   add-key-versus-weights shape (twelve total shapes across the
   eleven methods); a row's `weights` map is an override on
   `Tunables.DistanceWeights`, never a complete table; the
   `get_penalty` stub is wired through a loaded-source set built from
   `Tunables` (`metadataSourceSet`); and a compiled-regex option
   decodes from `{"regex", "ignorecase"}` into a `matcher.PatternOption`,
   never from a bare string.
7. **The `inclArtist` divergence**: `DistanceForTrack` takes
   `inclArtist` as a parameter, not a config read, because beets' two
   call sites disagree. The assignment solve (3d) calls
   `track_distance` with `inclArtist=false` for every track, while the
   final per-track distance a user sees passes `album_info.va`. The
   two objectives diverge exactly on VA albums; 3d's tie-mirroring
   task needs both call shapes.
8. **The `unidecode` pin check is split across two packages**:
   `matcher.UnidecodeVersion` (`translit.go:20`) is the exported stamp
   the embedded table carries; `conform.CheckUnidecodeVersion`
   (`defaults.go:67`) compares it against `corpus/manifest.json`'s
   pinned value, so a manifest bump reds the Go side too, not only the
   Python derivation path.
9. No row passed at 1e-12 without bit-identity. The near-miss list
   is empty across all 72 rows; there is nothing to name.
10. Tier-0 replay ran 0.02s standalone; the full `go test -count=1
    ./...` suite 13.4s. `bash scripts/check.sh` ran 1m45.6s and, on a
    second run minutes later with no tree change, 3m24.6s (both
    dominated by gitleaks' full-worktree scan, 99.6s and 3m13s
    respectively). The version-bump differential doc's next update
    should budget against the higher figure for the matcher and
    conform packages.

Two further findings sit outside the ten: the pointer ruling and the
Key-field ruling. `StringDist` and
`AddString` both take `*string` pairs where a nil pointer is Python's
`None` (Task 4's ruled API,
`StringDist(tun Tunables, a, b *string)`); every caller in `track_title.go`
and the album files passes a typed default's address, never a literal
nil, because `AlbumInfo`/`TrackInfo`/`ItemFields` fields are all plain
strings (Task 1), which is also the root of the `album_album`
divergence below. `trackComponent`/`albumComponent`'s `Key` field
sat unread from Task 5 through the start of Task 8; Task 8 gave it a
reader, `TestComponentKeysAreWeighted`, asserting every declared key
names an entry in `Tunables.DistanceWeights`, closing what would
otherwise have been dead struct data.

### Carried

- **To 3b (translation layer):** `addAlbumAlbum` passes
  `&likelies.Album`/`&info.Album`, both always non-nil, so a beets
  `album_info.album` of `None` scores `0.0` here where beets itself
  scores `string_dist("", None) = 1.0`. Root cause is Task 1's
  plain-string `AlbumInfo` fields; 3b needs to decide what an absent
  source field becomes before this divergence reaches a real case.
- **To 3b/3d (config validation):** `compileMediaPatterns` and
  `compileCountryPatterns` panic on a malformed user regex in
  `Preferred.Media`/`Preferred.Countries`, matching beets' own
  `re.compile` raising, but a wrapped error at config-load time is the
  better home once `internal/config` owns validation.
- **To 3c:** `dataSourcePenalty` falls back to
  `Tunables.DataSourceMismatchPenalty` (configured) where beets falls
  back to the class literal `0.5`; identical under shipped config,
  diverges only if `source_weight` is set.
- **Inert deviation, recorded and reviewer-verified, no action
  owed:** `MostCommonTags` counts every item with `len(items)` as the
  denominator, where beets' `if item` filter is a no-op for real
  `Item`s (no `__bool__`/`__len__`); `MostCommonTags` on an empty item
  set reports unanimity where beets asserts non-empty, unreachable
  today.
- `profile_hashes` (m1) and the response-status extension, 2d's own
  carry-forward, are unaffected by this pass and stay open. The
  adjudication batch was already fully ruled before this pass started:
  18 oracle-wrong, 3 oracle-correct, zero pending
  (`corpus/analysis/adjudications.json`, ruled Geoff 2026-09-04), and
  3a produced no new adjudication row. Rung 2's Tier C exclusions
  (`match.required`/`match.ignored`, `preferred.original_year`) are
  now the M8-declared gap documented above rather than an open
  unknown.

Budgets: ≈2.84M subagent tokens (T2 review 74K, batch 1 445K, batch 2
561K, batch 3 938K, `code-simplifier` 238K and its review 63K, the
close draft 193K, its prose review 63K, and the fold 261K) against the
6M starting ceiling, well under; conductor context on top. Attended
interaction points: zero. The one owner touch was a model-default
change mid-close, unrelated to the pass.

## Rung 2, pass 2d (File-level applied tags, the version-bump differential, and the whole-corpus README close): file_tags.json over 12 subjects, rung 8 unblocked (2026-09-04)

Four plan tasks executed sequentially against
`docs/superpowers/plans/2026-09-03-rung-2d-file-level.md` (revision 2).
Commits in order: `8d181fe` (Task 3, which ran first per the plan's
sequencing), then `4d38717` (Task 1), then `f5abba0` (Task 2), then
`412471f` plus `cc400ad` (Task 4's close and its same-session
forward-reference fix).

What landed: the template FLAC fixture and its provenance script
(`internal/harness/oracle/testdata/template.flac`, 42 bytes, silent,
untagged, authored via the pure-`mutagen` route with no `ffmpeg`/`flac`
dependency), `derive_file_tags.py` and its Go driver implementing the
write-then-independent-read-back contract with `BEETSDIR` isolation
proved rather than assumed, and `Manifest.MediaFile` pinning
`mediafile==0.17.0` at `schema_version` `3` with `Reconcile`'s `pinned`
tuple and `versionProbe` string both gaining their sixth component
(Task 1); `corpus/analysis/file_tags.json`, generated over a verified
twelve-subject set, its pure-Go gate-resident checker
(`internal/harness/analysis/file_tags.go`) wired into `corpus_test.go`
with three firing fixtures (an unresolved case id, a value key outside
the writable-set header, and — the fixture with real teeth — a row
carrying one of the nine read-only keys), and
`TestFileTagsLedgerMatchesTheCorpus` on the
`TestComponentTableReDerivesToItsCommittedTable` template (Task 2); the
version-bump differential job doc and the quarterly re-capture chore
doc, both standing operational docs beside `docs/bringup-runbook.md`,
the differential doc's wall-time section priced at the real 215-case
corpus with the file-tags line left as a named placeholder for this
task (Task 3); the whole-corpus case-inventory section (215 rows
across five stratum tables, built by a mechanical script pass over
each case's own `id`/`profile`/`eval_date` cross-referenced against
the family/task/adversarial/stock-config tables already committed,
never by re-reading a case body) and the rung 8 handoff statement,
both new `corpus/README.md` sections; the file-tags line written back
into the differential doc's placeholder (Task 4). Corpus total: still
**215 cases**, unchanged from 2c's close — 2d authors no case, captures
nothing, and edits no committed record, exactly as its own global
constraint required.

**The twelfth file-tags subject is a ruled substitution, and this
pass's own staleness catch.** The plan's Task 2 named
`real-decoy-beethoven-symphony-5` as the two-release sub-axis's
file-tags subject, drawn from the planning inventory dossier written
before 2c executed. That id was never captured: 2c's own Task 4 had
already proved the decoy unreachable inside beets' 5-result search
window with the axis's sole reserve spent, closing the axis at 7-of-8
(`corpus/README.md`, "The two-release sub-axis finding"). Task 2's
implementer caught this with a STOP rather than padding or trimming
the twelve-subject set, and the conductor ruled a substitution:
`real-decoy-mozart-requiem`, chosen from the axis's four other
committed decoy cases by eligibility (all four carry a non-empty
`applied_fields` record regardless of `action`, since
`match.apply_metadata()` runs whenever a candidate exists at all) and
then by the most distinctive applied-tag shape among the eligible rows
(24 items, two movement-numbering conventions, two distinct
`albumartist_credit` strings). Recorded with its eligibility evidence
in `corpus/README.md`'s file-tags section. **The standing lesson: a
plan drafted before a prior pass closes must re-verify its id lists
against the tree at execution time, not trust the dossier that
predated the pass's own findings.** This is not a new failure mode —
2c's own Task 17 alias-duplicate incident carries the sibling lesson
("check `corpus/cases/` before any capture") — but this pass's version
of it is a plan citing a specific id the immediately prior pass had
already retired, which a same-pass cross-reference (rather than a
later implementer's own tree check) could in principle have caught
earlier. It did not block the pass; the implementer's own verification
step, required by Task 2's contract regardless, is what caught it.

What the gate caught: nothing novel — `bash scripts/check.sh` and
`GOFLAGS=-count=1 make derive` stayed green through all four tasks,
and the pass's own fixtures (Task 1's silent-direction `Reconcile`
proof, Task 2's three gate-resident checker fixtures plus its
re-derive-catches-a-perturbation fixture) are what proved each new
check has teeth rather than the gate itself surfacing a defect in
existing code.

Determinism: two independent uncached (`GOFLAGS=-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree with the file-level derivation in the tree, 215
and 215 cases actually re-derived of 215 committed
(`TestCorpusReDerivesToItsCommittedRecords`'s own per-case subtests),
both runs exiting 0 with zero failures across every package, both
byte-identical to the committed records (corpus, ties, and component
ledgers, plus the twelve-subject file-tags ledger — `git status`
reported no diff under `corpus/` after either run). Wall time: run 1
980s (16m20s) total (corpus derive 462.94s, ties derive 503.90s,
component derive 0.75s, file-tags derive 0.59s); run 2 974s (16m14s)
total (corpus derive 541.20s, ties derive 420.06s, component derive
0.71s, file-tags derive 0.55s) — this is rung 3's budget line. 2d's
pair (980s, 974s) ran materially faster than 2c's own pair
(~1729.58s/1582.53s) on the identical 215-case corpus, roughly 40
percent faster; the spread is attributable to run-to-run variance in
request pacing rather than any corpus change, and the file-tags leg's
sub-second cost (0.55s-0.59s) is negligible under either figure. The
file-tags leg issues no MusicBrainz
request (twelve subjects, one Python process reading already-committed
`AppliedFields`), so it adds well under one second against the corpus
and ties legs' own multi-hundred-second cost; the measured figure
(0.55s-0.59s) is written into
`docs/corpus-version-bump-differential.md`'s previously-placeholder
row, alongside a new finding that doc records: the corpus and ties
legs' own M8-formula floor (257.6s each) is roughly half the real
measured wall time (462-541s and 420-504s respectively) — the formula
prices only the client's inter-request wait, not the Python/beets/
response-parsing overhead sitting on top of it, so a future
version-bump differential run should budget against the measured
figure, not the formula table.

What a later pass would be wrong to rediscover:

- **The rung 8 handoff statement lives in `corpus/README.md`** ("##
  Handoff to rung 8: what pure-Go FLAC tag writing may assume"),
  positioned immediately after "## The analyses" so rung 8's
  implementer finds it beside the evidence rather than in a fifth
  loose document. It states the two altitudes of expected applied
  tags (`Expected.AppliedFields` on every case, `file_tags.json` on
  twelve subjects), names the 69-key writable set as the only
  conformance target and the 9-key read-only remainder as never one,
  and states its own non-scope explicitly: the apply-metadata field
  mapping close read and the wazero-taglib spike are rung 8's own
  plan-opening work, not something 2d evaluated or prototyped.
- **2c's `profile_hashes` gap (m1) is still open**: `corpus/manifest.json`
  carries a `profile_hashes` entry per profile, but nothing in the gate
  or the derivation path computes or checks it for any profile yet.
  2d did not touch this; it is still scoped follow-on work, not
  something the manifest's own presence implies is built.
- **The response-status extension is still deliberately not taken.**
  Carried from 2c: the stub has no third outcome and the schema
  carries no response status, since no known-oracle-bug class probed
  so far needs a 404 or 503 to reproduce. Available to whichever later
  rung first has a case that genuinely needs it.
- **The Tier C exclusions and profile gaps 2b and 2c both carried,
  neither pass closed, carry forward again unchanged**: multi-source
  ordering and most of the `data_source` parameter space stay blocked
  by the single-source ruling; `match.required`/`match.ignored`
  candidate-dropping stays unprobed; `preferred.original_year: yes`
  stays unreachable under `estate`, `stock`, and `locale` alike (a
  fourth profile or a `locale.yaml` extension is still owed if a later
  rung wants this branch probed). 2d's own scope never touched any of
  these; nothing new closed here either.
- **The pending adjudication batch is still the launch gate's open
  item**, unchanged by this pass: nineteen rows across five hypotheses
  sit at `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md`).
  2d's own gate accepts them indefinitely, per its own "Owner
  touchpoints: none" scoping; it is the launch gate, later, that
  refuses an unadjudicated divergence. 2d produced no new
  adjudication row of its own.
- **The `schema_version` `2` -> `3` bump forced no re-freeze**, the one
  place this pass's schema event differs from 2b's: `Manifest` is
  corpus metadata rather than an `Expected` field, so `Manifest.MediaFile`
  landing needed no re-derivation of any committed case's own record,
  only the two-clean-run byte-identical proof above.

Carried:

- Rung 2 is complete at this pass's close: the smoke, structural,
  real-catalog, adversarial, and stock-config strata (215 cases), the
  file-level applied-tag ledger, and the two standing chore docs are
  all shipped. Rung 3 (the Go matcher) is what rung 2's completion
  unblocks, gated on Geoff reading its revision-2 spec before
  execution starts (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`).
- Every item 2c's own carry-forward left open (`profile_hashes`, the
  response-status extension, Tier C exclusions, the pending
  adjudication batch) stays open into rung 3 and beyond, restated
  above rather than re-derived.

Budgets: ≈1.6M subagent tokens (Tasks 3/1: 787k with the first
Task 2 STOP; the ruled resume of Tasks 2/4: 728k; code-simplifier:
94k) against the 2.5M starting ceiling, well under. Attended
interaction points: zero — the twelfth-subject STOP resolved to the
conductor, exactly as the plan's owner-touchpoints section required.

## Rung 2, pass 2c (Real-catalog and adversarial strata): 147 new cases, six oracle-bug hypotheses, first adjudication batch (2026-09-03)

Twenty plan tasks executed across a single sequential pass-execute
chain against
`docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md`
(revision 2). Commits `3bbdfbd` → `890bb70` (Task 20 close at `f5bb1bd`; the
locale re-base and 19b prose fix at `890bb70`).

What landed: the adjudication and known-oracle-bug ledger formats
(Task 1); the `harness seed-items` projection and `--facts` tools
(Task 2); the real-catalog selection dossier, 96 primary rows plus 12
reserve rows across seven axes (Task 3); twelve capture batches (Tasks
4-15) landing 94 real-catalog primary cases against real MusicBrainz
document shapes (two two-release sub-axis rows lost to the mechanism's
own escape hatch, real captures over MB-noise, non-Latin script,
classical/compilation, genre/era spread, reissue/remaster, and
VA/box-set releases, one `ignore_video_tracks` finding); the
adversarial stratum (Tasks 16-18), synthetic cases reproducing two
known string-heuristic beets defects (the `"ft"` word-boundary false
positive and the parenthetical/remix near-duplicate discount, ten
cases, both moving strong/apply to medium/skip) plus six
per-pair track-assignment rejection-gap cases (the album verdict never
moves even while `assignment`/`applied_fields.title` are provably
wrong) plus the alias-mis-selection hypothesis probed against two of
Task 4's own real-nonlatin cases, plus Task 18's `locale` profile and
its A/B regression-check case pair for `#6358`; the stock-config
stratum (Tasks 19a/19b), 36 siblings (28 mandatory + 8 supplementary)
proving zero verdict-level divergence between `estate` and `stock`
across every real-catalog axis, with the divergence landing only in
`applied_fields` (`year` for reissues, `tracktotal` for box sets).
Corpus total: **215 cases** (7 smoke + 61 structural + 94 real-catalog
primary + 17 adversarial + 36 stock-config), read off the tree
directly. Determinism: two independent uncached (`-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree, 215 and 215 cases actually re-derived of 215
committed (`TestCorpusReDerivesToItsCommittedRecords`'s own per-case
subtests, the count that proves a case was actually re-derived rather
than skipped for want of a committed record), both runs exiting 0 with
zero failures across every package, both byte-identical to the
committed records. Wall time: run 1 1729.58s (28m50s) total (corpus
derive 927.91s, ties derive 786.04s, component derive 1.22s); run 2
1582.53s (26m23s) total (corpus derive 812.66s, ties derive 754.48s,
component derive 1.27s) — this is rung 3's budget line, exactly as
2b's was 2c's own (2b's own pair ran ~2m13s each on 68 cases; 2c's
real-catalog and box-set growth moved that budget line by roughly
13x). The heaviest single cases in either derive phase are the
box-set axis's largest releases: ties derive's own re-solve-per-pair
algorithm (`internal/harness/oracle/derive_ties.py`) pays for
`real-boxset-rachmaninov-complete`'s 470 tracks and its stock twin
individually, each taking on the order of a minute alone, the same
box sets that dominate the byte-growth figure below. Corpus byte
growth (m7): the committed corpus held 4.0 MB across 114 response
files before this pass; it now holds **81.3 MB across 351 response
files**, well past the plan's projected 10-15 MB
landing point. The overshoot is a response-*size* story, not a
response-*count* one (351 files is close to what the plan's own count
math expected): a handful of the VA/box-set axis's largest releases
serialize their entire tracklist inline in one WS/2 document
(`real-boxset-rachmaninov-complete`, 470 tracks, 9.1 MB alone;
`real-boxset-scriabin-complete`, 321 tracks, 3.1 MB;
`real-boxset-mahler-complete`, 133 tracks, 1.7 MB), and Task 19a's
mandatory-sibling `cp -r` rule doubles every one of them for its stock
twin.

What the gate caught: nothing new at Task 20 itself (README-only and
analysis-doc changes, no code); across the pass, the recurring catch
was a case's own `claims.json` failing to notice a degraded record
(the claims-checker's whole design point) and, once, a stratification
undercount in Task 19b's own README prose (fixed same-pass, commit
`9915907`).

What a later pass would be wrong to rediscover:

- **File-level derivation is 2d's Task 13, unchanged in scope**:
  `internal/harness/oracle/derive_file_tags.py` and its Go driver,
  reading back independently through `mediafile.MediaFile(path)` after
  `item.write()`, never trusting the write call's own return value,
  over 12 subject cases (5 from the committed corpus, 7 flagged by
  this pass's dossier).
- **The audio-fixture route is ruled** (M4): vendor one small static
  template FLAC blob under `internal/harness/oracle/testdata/`,
  committed, with its provenance and byte size recorded in the README.
  `make derive` gains no binary dependency; `ffmpeg`/`flac` are not
  adopted as prerequisites.
- **`derive_file_tags.py` is a second, independent entry point** (M5),
  shaped like `derive_component.py`: it reads a committed record's
  `AppliedFields`, builds beets items from those fields, allocates its
  own scratch directory, writes and reads back there, and is never
  called from `oracle.Derive`. The invariant: no item in the corpus
  derivation path ever gains a non-empty `path`, proved by `make
  derive` re-deriving all committed records byte-identically after 2d's
  change.
- **`file_tags.json` needs an oracle-tagged re-derive-and-diff test**
  (M3), `TestFileTagsLedgerMatchesTheCorpus`, matching
  `TestTieLedgerMatchesTheCorpus`/`TestComponentTableReDerivesToItsCommittedTable`'s
  precedent, plus the `-update-ledgers` regeneration path documented in
  `corpus/README.md`. A gate-resident structural check alone (case ids
  resolve, pointers resolve) is not enough; nothing else would catch a
  `mediafile` version bump moving the tag shapes.
- **`Reconcile` lives in `internal/harness/oracle/reconcile.go`, not
  `oracle.go`, and its `versionProbe` tuple names five components
  today** (m2: beets, unidecode, jellyfish, lap, numpy), not four. 2d's
  Task 13 must edit `reconcile.go` and `versionProbe` explicitly when
  it adds `mediafile` to `Manifest`, or a reconcile silently compares
  against a missing key.
- **The version-bump differential and quarterly re-capture chore docs
  are unwritten**, deferred with Task 13. The differential doc's
  wall-time section must use the M8 inter-request-floor formula
  (`1.04s + 0.25s × (requests − 1)`) at the corpus's real size once 2d
  closes, not a flat per-case figure.
- **The whole-corpus README rewrite happens again at 2d's close**,
  once, to add the file-tags section; this pass's rewrite is not
  wasted work, since 2d's version extends it rather than replacing it.
- **The response-status extension was deliberately not taken this
  pass**: the stub has no third outcome and the schema carries no
  response status, because all four string/assignment/alias
  known-oracle-bug classes this pass probed are distance or selection
  defects, not transport defects — none needs a 404 or 503 to
  reproduce. It carries forward as a named non-adoption, available to
  whichever later rung first has a case that genuinely needs it.
- **The `profile_hashes` gap (m1)**, distinct from the 2d file-tags
  items above: `corpus/manifest.json` gained a `locale` entry under
  `profile_hashes` at Task 18, but nothing in the gate or the
  derivation path computes or checks it yet for *any* profile. A
  profile edit does not fail loudly the way a beets version pin
  already does; this is scoped follow-on work, not something the
  table's presence implies is built.
- **Tier C exclusions and profile gaps 2b carried that 2c did not
  close**: multi-source ordering and most of the `data_source`
  parameter space stay blocked by the single-source ruling (unchanged
  from 2b); `match.required`/`match.ignored` candidate-dropping stays
  unprobed (same shape, still open); `preferred.original_year: yes`
  stays unreachable under all three profiles now committed (`estate`,
  `stock`, and Task 18's own `locale`, which only toggles
  `import.languages`/`import.ignored_alias_types`) — a fourth profile
  or an extension to `locale.yaml` is still owed if a later rung wants
  this branch probed.
- **The alias-duplicate incident (Task 17) and its standing rule**: an
  earlier pass of Task 17 captured four byte-identical `adv-alias-*`
  cases duplicating releases the corpus had already committed under
  Task 4's `real-nonlatin-*` rows, because case selection never
  checked the committed tree first. The conductor's escalation ruling
  folded the four duplicates back into their existing twins (commits
  `e805540`, `c685894`); nothing was lost, but the rework was pure
  waste. **The rule a later pass must inherit: check `corpus/cases/`
  before any capture**, real or synthetic, that might already exist
  under a different id.
- **The `_preferred_alias` unreachability finding and Task 18's
  resolution of it**: Task 17's alias-mis-selection probe found
  `_preferred_alias` (`beetsplug/musicbrainz.py:128-155`) structurally
  unreachable under both `estate` and `stock`, since both leave
  `import.languages` empty, so its generator yields nothing regardless
  of what a release document carries. This was left as a scoping
  finding, not a reproduction, and deferred to Task 18, which added the
  `locale` profile and an A/B case pair
  (`real-nonlatin-utada-first-love` / `-locale`) to probe it directly.
  Task 18's first cut based `locale.yaml` on `stock.yaml` (as the plan
  itself prescribed), which the review caught as a confound: the pair
  carried the whole estate/stock delta besides the locale keys. The
  conductor ruled the clean experiment and the profile was re-based
  same pass (commit `890bb70`): `locale.yaml` is now `estate.yaml`
  plus exactly `import.languages: [ja]` and
  `import.ignored_alias_types: []`, so every divergence in the frozen
  pair is attributable by construction. The result: the mechanism *is*
  reachable and *does* fire once `import.languages` names a matching
  locale, moving the four sort-name fields it owns (`artist_sort`,
  `albumartist_sort`, `artists_sort`, `albumartists_sort`) on all 12
  items, plus one genuine bonus finding reported verbatim: the
  MusicBrainz "Various Artists" *entity* itself carries a primary
  `locale=ja` alias, so `_preferred_alias` also fires for two
  VA-credited losing candidates and moves their scoring inputs
  slightly (no verdict effect). Everything else in the pair is
  byte-identical — evidence against `#6358`'s report, pending Geoff's
  own ruling (`bug-locale-alias-ignored`, adjudication batch 1).

Carried:

- **The pending adjudication batch is the launch gate's open item.**
  Nineteen rows across five hypotheses sit at
  `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md` is
  the artifact prepared for Geoff's ruling). This pass's own gate
  accepts them indefinitely; it is the **launch gate**, a different
  gate at a later time, that refuses an unadjudicated divergence.
  `bug-threshold-boundary` (`#5458`) is the corpus's one hypothesis
  already ruled (`oracle-correct`, Geoff, 2026-09-03) and is not part
  of the pending batch.
- `corpus/analysis/divergences.md` stays empty at this pass's close,
  reviewed and carried forward: no case across the real-catalog,
  adversarial, or stock-config strata contradicted a specific beets
  test assertion it was modeled on (findings of that shape instead
  became adjudication-batch rows or plain README prose).
- The two-release sub-axis closed at 4-of-6 primary rows delivered (2
  discarded via the mechanism's own escape hatch, the sole reserve
  spent by Task 4); a later two-release row anywhere in the corpus
  should budget for the same WS/2 search-ranking risk with no reserve
  behind it.
- Every item from 2b's own carry-forward that 2c did not separately
  close above (Tier C, `match.required`/`match.ignored`,
  `preferred.original_year`) stays open into 2d and beyond.

Budgets: ≈8.2M subagent tokens total (≈6.5M through Task 17's fold,
1.4M for Tasks 18-20's chain, 0.27M for the close fixes) against the
8M post-Task-3 re-derived ceiling (12M was the plan's pre-re-derivation
starting figure; finishing past the 80% flag approved by Geoff at the
flag, 2026-09-03). Attended interaction points: one mid-pass (the
combined 80%-flag question); the adjudication batch is delivered, not
yet spent.

## Rung 2, pass 2b (Structural stratum): analyses, synth, 61 synthetic cases (2026-09-02/03)

Fourteen plan tasks plus one staleness pre-task, executed overnight
across seven pass-execute workflows against
`docs/superpowers/plans/2026-09-02-rung-2b-structural-stratum.md`
(revision 2). Commits cec3c12 → 2219b19. All tasks accepted; five
needed one fix round; four escalations, each ruled by the conductor
and recorded in
`research/2026-09-02-rung-2b-staleness-corrections.md`.

What landed: the machinery half (threshold/weight echo with schema v2
re-freeze; `harness freeze` with `Options.Dir`; `harness synth` and
the authoring loop; the proximity flag, tie detector, per-case claims
checker, and 72-row component derivation table; the shared
`mbrequest` package), then the 61-case structural stratum across all
15 shape families, every record derived through real pinned beets,
plus the claims schema's zero-candidate mode and `penalty_keys_absent`
field. Corpus total: 68 cases, all re-deriving byte-exact across two
independent uncached runs (~2m13s each: corpus 70.5s, ties 54s,
component 0.7s; this is 2c's budget line).

What a later pass would be wrong to rediscover: the staleness review
pattern paid for itself (a fresh-context read of the case tasks
against the machinery as shipped found five blockers before any
dispatch; plan staleness remains the dominant defect class); the
"combined" completeness shape is unexpressible (`lapjv`'s leftover
set is one-sided); the VA-misflag has NO record-level observable
under the empty-search convention, so its witness is the frozen
request log's `arid:` query, which `make derive` mechanically
enforces; the response hand-edit escape hatch (edit generated
responses BEFORE first freeze, named in the README row) is the
sanctioned route for shapes `ReleaseSpec` cannot express; a
sub-second `make derive` is a cache hit and proves nothing, so force
`-count=1` for evidence.

Carried: two smoke cases tied (pinned; rung 3's solver must mirror
lapjv tie-breaking); family 15 parks cases on thresholds with
acknowledged proximity rows; the candidate-absent country direction
documented as declined; claims formatting drift cosmetic. 2c's
inventory is unchanged from the 2a plan's split section, plus the
determinism budget line above.

Budgets: ≈5.8M subagent tokens against the 6M ceiling (Geoff's 12M
overnight grant untouched); human interaction: two batched bedtime
questions (ceiling, 2c scope), zero mid-execution.

## Rung 2, pass 2a (Cassette substrate and capture): schema, stub, shim, capture tool, smoke corpus (2026-09-02)

Executed under two pass-execute workflows (wf_5ed5b773, wf_fe30d704)
plus one escalation follow-up, against
`docs/superpowers/plans/2026-09-02-rung-2a-cassette-substrate.md`
(revision 2). Commits 79420d3 → 69a1804. Five plan tasks, all accepted;
one fix round (task 5); one escalation, ruled in-scope and closed by
follow-up commit 6d45c28.

What landed: `internal/harness/cassette` (the frozen schema, verbatim
response-file store, byte-wise canonicalization rule);
`internal/harness/stub` (replay server with a miss list) and
`internal/harness/oracle` (adapter proven against a fake shim, then the
real `derive_oracle.py` driving pinned beets 2.13.1 in a uv venv, with
internal self-verifying date pinning and a raw-stdout determinism
proof); the two config profiles with their documented transform;
`cmd/harness` (capture tool: `--mbid` seeding with maximal `inc=` and
the pseudo-release double fetch, `--case` miss consumption, 1 req/s
identified wire discipline, contact refusal); and a seven-case live
smoke corpus (clean-match, should-ask, should-reject, various-artists,
multi-disc, non-latin pseudo-release with its transl-tracklisting
follow-up, year-penalty), every record derived through the real shim,
all seven re-derivable from the committed tree
(`TestCorpusReDerivesToItsCommittedRecords`) with the date-pinning
proof riding the real year-penalty case. Gate grew ruff and gitleaks
legs; `make derive` runs the oracle-tagged tests.

What a later pass would be wrong to rediscover: the `inc=` join is
`%2B` on the wire, never a literal `+` (requests percent-escapes it;
every live case missed its primary lookup until the capture tool's
canonicalization matched, and the original cassette doc's "literal +"
wire fact was corrected in 6d45c28; do not "fix" it back); the stub
must key on `r.URL.EscapedPath()`, since `r.URL.Path` arrives
percent-decoded; the miss loop's contract is `<case dir>/misses.json`,
written by `Derive`, consumed by `harness capture --case`; a
strong ID match legally orphans search captures into `Extra` (recorded
fixed point, not a stable request set).

Carried to 2b/2c: three non-blocking hardenings on
`corpus_derive_test.go` (skip-all-green residual hole, unguarded
`Candidates[0]`, redundant equality branch); the line-pinned
`.gitleaksignore` fingerprint; git-history gitleaks scanning (deferred
as a heavier job); response-number gaps in two cases (documented,
harmless).

Budgets: ≈1.75M subagent tokens against the 5M ceiling, plus conductor
context. Human interaction: one batched ask (the MusicBrainz contact,
minted as `DUBPLATE_HARNESS_CONTACT` in `~/.bashrc` via dotfiles
3a1c551; registration declined, anonymous 1 req/s stands).

## Rung 1 (Foundation): the repo pivot, config core, check/doctor (2026-09-02)

The greenfield build's first rung, run under the pass-execute workflow
(implementer/diff-reviewer/gate chain per task, owner opt-in). Landed
against `docs/superpowers/plans/2026-09-02-rung-1-foundation.md` and the
ratified build spec (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`,
revision 2.1): the three Open rulings that spec's review round raised
(shadow floor, oracle-bug policy, stock-config gating) were all resolved
by Geoff the same day, closing the item STATUS carried as "Open rulings
1-3 await Geoff."

Five tasks: the pivot commit tagged `pre-greenfield` and moved `cmd/`,
`internal/`, `web/`, and `e2e/` under `_salvage/`, freezing them
read-only, and stood up `cmd/dubplate-server` as the module's only Go
package; the ops audit removed `config/filebrowser.yaml` as an orphan
and made `scripts/deploy.sh`'s binary step additive (rsyncs to the new
`/usr/local/bin/dubplate-server` path, never touching the incumbent
`/usr/local/bin/dubplate` the four guardian units still exec); the
config core (`internal/config`) implemented the derived wall as code:
owned, reserved-for-a-rung, and superseded key classes, plus a
threshold registry, all contract-tested from a stub before
implementation; `check`
and `doctor` scaffolds gave the wall an observable surface with a pinned
line format; this ledger close reconciled ROADMAP, CLAUDE.md, and
`.claude/settings.json` against the pivoted tree.

What the gates caught and a later rung must not rediscover:
`scripts/vale-comments.sh`'s file-list union needed an explicit
`_salvage/` exclusion, because git's `*.go` pathspec matches at any
depth and would otherwise re-admit all 76 salvage Go files into the
prose-comment scan; `go mod tidy` pruned the module to cobra's set once
the salvage tree's other dependencies left the build; the additive
deploy boundary (new path beside the untouched incumbent binary) is the
mechanism that lets rung 1 land without any box write, and every rung
before the launch cutover (rung 16) must preserve it. No box access
occurred this rung, per the spec's box-safety section.

Post-close fix trail: task 4's review escalated a silent-error defect
(SilenceErrors set with no stderr print in main). Three commits closed
it under review—a531d69 (main prints the error once; check stops
double-printing), 1f3ef3f (doctor returns "configuration is
incomplete" instead of writing the raw load error), baac9f4 (the
no-usage-block contract gains a live test, mutation-proved)—and
db3d11b landed the conductor-dispatched code-simplifier pass. Two more
lessons for later rungs: cobra's usage writer falls back to os.Stderr
when only SetErr is wired, so command tests must wire SetOut to
observe a leaked usage block; and doctor's row expectations must
derive from config's own tables (owned, reserved, Thresholds()),
never a literal count. Budget: about 1.0M subagent tokens for the
workflow (12 agents, five tasks) plus about 0.3M across the fix and
verify dispatches, against the 5M ceiling; zero attended interactions
during execution. Verdict trail: T1-T3 accept first-pass; T5 accept
after one fix round; T4 escalate plus two fix rounds.

## Uploader Pass B close + the greenfield pivot (2026-09-01/02)

Pass B built and hardened the uploader web app end to end, then Geoff
superseded its launch with the greenfield dubplate-server ruling. What
landed in code (all on master, salvage tier): U9-U11 web app; the U13
review workflow (57 agents, 45 confirmed serious findings from 5
reviewers with adversarial verification) and its complete fold F1-F8
(TagProbe built and wired -- the probe never existed and no main-lane
upload could reach the library; serve security majors; reject-copy
contract; semantic theme tokens; App.svelte structure; component fixes;
saturated palette + soft surfaces; state-text lightness + rejected
CTA); a three-round visual-fidelity chain ending PASS; code-simplifier
consolidation (AlertCard, Store access discipline). Deployed to the box
2026-09-02 with "907 Music" minted, then wound down the same day.

What the gates caught that later work must not rediscover: DaisyUI 5
theme blocks define NO semantic color defaults for custom themes
(navidrome-* themes shipped without --color-error et al. and every
status surface rendered gray); DaisyUI 5 .alert is display:grid (flex
utilities are inert; single wrapper child required); -soft variants
color text with the raw token, so dark themes need Material-400-class
state tokens; the beets probe cannot use --pretend (it skips MusicBrainz
entirely) -- the real dry-run shape is in internal/beets/probe.go; the
box serve unit needed AF_NETLINK debate but not the family (glibc falls
back); find -newermt with a bare relative string fails silently in
guard scripts (use -mmin).

The pivot: Geoff's device check hit the shared-Access-org ASC branding
on the login page (a page never rendered before asking him -- now the
see-it-before-geoff-tests-it memory). The fix path escalated by
ruling into: magic-link login + Navidrome SSO (spec
2026-09-02-magic-link-login-design.md; two adversarial reviews in
docs/superpowers/research/2026-09-02-login-spec-reviews/ found 6
security blockers + 3 universality criticals -- all now birth
requirements), then the full greenfield ruling set
(docs/superpowers/research/2026-09-02-greenfield-server-rulings.md).
Interim estate wound down 2026-09-02: FileBrowser gone (container,
compose, env, DNS, Access app), uploader serve + nudge units disabled,
uploader tunnel rule + both Access apps deleted; box runs Navidrome +
cloudflared + guardian timers only. The 37 deferred U13 minors live in
docs/superpowers/research/2026-09-01-u13-findings.json and matter only
to salvaged components. The dubplate-organizer initiative spec was
brainstormed and four-review hardened the same nights (80f0f5f + later
amendments); the organizer spec + rulings doc are the greenfield
build's inputs.

## Uploader Pass A close: U12a live, gates 1–2 passed (2026-09-01)

The uploader went live at https://music.907.life/upload behind a
path-scoped Access app (Geoff relocated it mid-sitting from its own
hostname; the AUD survived the domain edit, so the minted secret held).
serve mounts its whole tree under /upload via StripPrefix with tusd's
BasePath threaded through Options.TusBasePath, because tusd bakes the
configured base verbatim into the absolute Location clients resume
against. Gates 1–2 were measured through the full production path by
logging in through Access with an emailed code read over JMAP and
driving fetch from the authenticated page: SSE heartbeats at exact 15s
cadence with no edge buffering, Last-Event-ID resume in 120ms, tus
offset persistence through an aborted PATCH. The gate found one defect
(tusd stamping http:// behind the tunnel; fixed same hour with
RespectForwardedHeaders).

What a later pass would be wrong to rediscover: Docker's host-gateway
magic resolves to docker0's gateway, never a named bridge's, so the
compose alias pins 172.30.0.1 literally; a compose network recreate
reattaches live containers WITHOUT service aliases (502'd Navidrome
until force-recreate; now a vps-conventions row); Access service tokens
pass Access but die at serve's roster-gated JWT check by design, so
machine verification goes through the email-code flow; the Access login
page wears the ASC org's default branding (Pass B fixes the app's own
name/logo). Pass B standing inputs live in
docs/superpowers/research/passb-{stack-brief,copy-deck}.md; the copy
deck carries seven pre-U9 revision items.

## Go rewrite close: box cutover and estate rename, T12–T14 (2026-08-31)

Geoff's no-users ruling collapsed the plan's staged migration (additive
env window, per-subcommand flips, soak-gated T13) into one evening: the
estate has a library worth protecting (backed up, verified) but no users
a botched cutover could hurt, so the windows bought safety nobody
needed. One continuous run: T12–T14 repo work through the chain (two
review rounds each; the reviewers caught a ReadWritePaths grant that
would have failed every backup — fixed by clearing staging CONTENTS so
the leaf grant holds — and a compose project-name float that would have
orphaned every container, fixed by pinning `name: musicbox`); one
deploy shipping the binary, nine units, and the 36-name
single-generation env at /etc/dubplate; doctor fully green first try;
all four jobs green on supervised first runs; old units/scripts/paths
removed from the box; pings checks re-registered as dubplate-* inside
the grace window (no false alert); `~/Projects/musicbox` →
`~/Projects/dubplate`. The soak runs post-hoc: overnight timers prove
unattended operation with the dead-man checks as witness.

What a later pass would be wrong to rediscover:

- Compose derives its project name from --project-directory's basename;
  moving the directory without pinning `name:` makes `up -d` orphan
  every running container and start duplicates. The pin carries a
  never-remove-without-migration comment in compose.yaml.
- A ReadWritePaths LEAF grant cannot survive os.RemoveAll on the
  granted directory itself (needs parent write). dubplate-backup clears
  contents instead; the unit comment carries the contract.
- Kept identifiers (Hetzner names, tunnel, compose project,
  /var/lib/musicbox, /srv/music) are documented in README; the estate
  rename is complete EXCEPT those, deliberately.
- `Paths.RunDir`/`DUBPLATE_RUN_DIR` is required by config.Load but read
  by nothing — dead config, escalated to the uploader pass's U1 rather
  than patched around the frozen inventory.
- Slug swap ordering that avoids false dead-man alerts: add new checks,
  re-mint URLs, redeploy env, prove one ping, THEN remove old slugs —
  all inside the shortest check's period+grace.

## Go rewrite, repo phase T1–T11 (2026-08-31)

The shell layer's replacement landed as one `dubplate` binary: sixteen
internal packages plus cmd and an e2e harness with stub beet/rclone/
sqlite3, all through the implementer/opus-diff-reviewer/gate chain via
two pass-execute workflows (wf_edc843fb, wf_048446d1) plus follow-up
dispatches. All eleven tasks accepted; consolidation added the
code-simplifier sweep, poplar's vale comment-prose gate (vendored
overlay, hard-fail), and tidy-check. Commits 6ec7aab → 4489c99. Spend
≈4.3M of the raised 5.5M ceiling (original 4M overrun approved by Geoff
at the 3.2M checkpoint); human interaction: plan approval (pre-session),
one combined overage/SMTP question, one per-step-sanction standing
grant. The chain caught real defects at every level: a config env-scan
panic, a vacuous TLS assertion, ping-fail exit semantics, two importer
shell-parity divergences, and a contract-wording regression (T5).

What a later pass would be wrong to rediscover:

- An implementer agent can DIE mid-task (API drop) leaving compiling but
  untested, unreviewed code; the workflow journal records nothing for it
  and neighbor tasks inherit red gates that look like their own
  failures. Signature: a task with no journal record + innocent
  escalations citing out-of-scope lint. Recovery: complete + review the
  orphan task explicitly; re-review the neighbors after.
- Dispatches carrying a reviewer's verbatim LOCATION/FINDING/FIX text as
  acceptance criteria went through at zero fix rounds; paraphrased
  criteria averaged one. Pre-extract findings, never summarize them.
- Rejection wording is per-extension via a path-aware builder inside
  RenderEmail (Friendly() cannot carry the extension; its signature is
  contract). The shell's ungrammatical "an WAV" is pinned deliberately.
- `internal/beets` carries `subsonicClient = "musicbox"` matching the
  live shell; it flips in T14's rename sweep, never earlier.
- review.Filer's dirs are direct destinations, not tree roots
  (documented at review.go:22): importer builds two differently
  configured Filers, one for filing, one for sweeping.
- The e2e ENOTDIR blocked-path fixture (a plain file where a dir is
  expected) is deliberate: chmod fixtures pass trivially under root.

A separate agent executed the bring-up runbook against the live box; the
close-out session verified steps 5–7 read-only from the box and committed
the outcomes. Seed result: nine ZIPs in, seven albums imported (Angel Du$t,
Bully, Cheekface, De La Soul, Nightmares on Wax, TV Girl, Wax Tailor — 7
albums / 105 items in the beets DB), two review-queued with reasons
(Aim "Hinterland (Remastered)", illuminati hotties "POWER") plus a TV Girl
zine PDF routed to review. Post-seed backup succeeded 18:06 UTC:
`library/` in R2 at 112 objects / 2.794 GiB, `state/` holding both DB
snapshots. Monitoring truth: import and backup success pings green; the
failure-ping path is confirmed dead on the box (curl 404 in both
`musicbox-ping-failure@` journals — the known pre-wave GET-/fail bug) and
the disk/navidrome check timers were never enabled by the pre-wave deploy;
both fixed by the interim deploy this entry accompanies.

What a later pass would be wrong to rediscover:

- Contributor ZIPs extract FLAT (loose `Artist - Album - NN Track.flac`),
  so beets needs `group_albums: yes`; without it the whole seed read as
  one 135-track candidate and skipped.
- Since beets 2.4, MusicBrainz is a metadata-source plugin: an explicit
  `plugins:` line without `musicbrainz` leaves the autotagger with zero
  candidate sources and every album skips.
- Hetzner Cloud blocks outbound 25 and 465 by default on new accounts;
  Fastmail submission works on 587 STARTTLS from the box.
- `safe_extract` needs bsdtar on the host; it was missing from cloud-init
  packages and broke the first seeded import (hand-installed on the box,
  cloud-init now carries it).

## Architecture refinement pass (2026-08-30/31)

Adversarial review of the estate as built, then a fix wave. Phase A: nine
finder lenses plus an opus verify wave over both repos, the spec chain, and
live Cloudflare state — 72 findings, 40 confirmed, 12 endorsed, 0 refuted
(ledger: ~/.dotfiles/docs/superpowers/plans/2026-08-30-musicbox-arch-refinement-findings.json).
Fix wave: 11 tasks + lock follow-ups through implementer/diff-reviewer/gate
chains, all accepted (commit 087df6c); pings hardened in its own repo
(40fdfa7). Spend ~6.2M tokens against a 6M ceiling (3% over, driven by an
uploader design brainstorm folded into the same session); human
interaction: one combined triage question, approvals otherwise.

What a later pass would be wrong to rediscover:

- A lock file in a systemd RuntimeDirectory is a lifecycle bug (deleted at
  unit stop, unlinked-inode races across units). Locks live in
  /var/lib/musicbox; per-run scratch keeps /run. Doctrine now in the
  vps-conventions skill — invoke it before touching units.
- ping_hc fail is POST; the Worker 404s GET /fail. Every failure ping ever
  sent before this pass was dropped.
- bash cannot re-arm errexit inside a command-substitution subshell; the
  M3 first attempt guarded notify() with a subshell set -e that did
  nothing. Best-effort is enforced by if-guarding at the definition.
- The beets database is the only import success record (tag log = skips
  only; beets renames everything). The uploader design builds on this.
- Review standard that caught real bugs: verify test non-vacuity by
  mutation (M10's reviewer), and re-run implementer claims yourself
  (M2b's reviewer caught a misdescribed unit).
- The estate renamed musicbox → dubplate (2026-08-31); repo/paths converge
  at rewrite T14. The rewrite, uploader, and packaging specs were authored
  and adversarially reviewed this same session.

Carried out of the pass: SELinux one-time check on /var/lib/musicbox at
first deploy; provision.sh's ssh_cidrs_json exits only its substitution
subshell on malformed input (diagnostic degrades to a jq error); pings
remove-check has no --force for a permanently orphaned route record.

## Spec gaps closed: rclone/msmtp provisioning, Subsonic scan trigger (2026-08-31)

Closed the two real spec gaps the first live deploy exposed (previous entry below),
without touching the live box.

- **`rclone`/`msmtp` were never in `cloud-init/user-data.yaml` at all.** Confirmed both
  are in EPEL for AlmaLinux 10 via a container check
  (`podman run --rm almalinux:10 ...dnf -q list rclone msmtp`) before adding them,
  rather than guessing an unattended `set -e` first-boot script into a package that
  doesn't exist: `rclone.x86_64 1.74.3-1.el10_2` and `msmtp.x86_64 1.8.25-6.el10_1`,
  both from the `epel` repo alone, no RPM Fusion or CRB dependency. Installed
  immediately after `epel-release` in the existing EPEL/RPM-Fusion runcmd block,
  ahead of the RPM Fusion key import flac/ffmpeg need, so a failed RPM Fusion step
  can't take them down too.
- **`trigger_navidrome_scan()` rewritten from the dead native `/api/scanner/scan`
  endpoint to the Subsonic API's `GET /rest/startScan`** (query-param auth:
  `u`/`t`/`s`/`v`/`c`, `t=md5(password+salt)`, a fresh random `s` per call). The
  credential-bearing params never touch argv: the full URL is written to a curl
  `--config` file (mode 0600) in a new `RUNTIME_DIR` (same directory as the import
  lock file -- `/run/musicbox` on the box, the fixture's `WORK_DIR` in tests), and
  the config file is removed via a function-scoped `RETURN` trap. Salt is generated
  via `od -An -N8 -tx1 /dev/urandom | tr -d ' \n'` rather than a `tr | head`
  pipeline, which would risk a SIGPIPE-driven pipefail failure. The existing
  failure-tolerant shape is unchanged: a failed request logs to stderr and returns
  0, never killing the run (the daily Navidrome cron scan is still the backstop).
  `tests/music-import.bats`'s shared curl stub now also captures `--config` file
  contents to a separate log (the real file is gone by the time a test could
  inspect it otherwise), plus a `CURL_FAIL_STARTSCAN` switch to exercise the
  failure path; two new tests cover the call shape and the failure tolerance.

Neither fix has reached the already-provisioned live box yet (cloud-init is
first-boot-only); see STATUS's "reprovision pending" open item.

## Task 4/6: first live deploy (2026-08-31)

`scripts/deploy.sh --check` then a real run against the provisioned box. Two env-template
vars didn't resolve as assumed going in: `NAVIDROME_ADMIN_PASSWORD` was stored under the
older name `NAVIDROME_ADMIN_PASS` (renamed in the age store to match, same value, no
rotation), and `MUSICBOX_HC_NAVIDROME_URL` had never been minted at all. Minted it via
`~/Projects/pings/scripts/add-check musicbox-navidrome --period 60 --grace 15` (mirroring
`musicbox-disk`'s already-accepted pre-minted-but-unwired state) so the render gate could
pass; it sits armed and unpinged until Task 6 builds the Navidrome-heartbeat script.

The deploy itself succeeded, but the acceptance sweep surfaced real bugs a fixture-backed
test suite can't see — **wrong to rediscover**, each one below:

1. **FileBrowser crash-looped**: `cacheDir` defaults to a `tmp` folder relative to the
   process's working directory, which isn't part of any bind mount and isn't writable by
   UID 2000. Fixed by pointing `server.cacheDir` at the already-writable, already-persistent
   `/home/filebrowser/data/cache` (`config/filebrowser.yaml`).
2. **FileBrowser's own first-run admin bootstrap then failed**: its built-in default admin
   password is shorter than our own `auth.methods.password.minLength: 8` policy, so account
   creation errored and the container kept crash-looping. Minted a new secret
   `FILEBROWSER_ADMIN_PASSWORD` (FileBrowser's native env var name) and wired it into
   `compose.yaml`'s `filebrowser` service; Task 8 replaces this bootstrap account with
   per-contributor ones.
3. **music-import.service failed on `/run/lock/musicbox-import.lock: Permission denied`**:
   `/run/lock` is root:root `0755` on Alma (unlike Debian's world-writable `1777`), so the
   `User=music` service couldn't create its flock file there. Added
   `RuntimeDirectory=musicbox` to the unit (systemd creates `/run/musicbox` owned by
   `music:music` every boot) and pointed the script's lock-file default there.
4. **music-import then failed on `curl: (7) ... Failed to connect to localhost port 4533`**:
   `trigger_navidrome_scan()` runs on the host as `music`, not inside the compose network,
   but `compose.yaml`'s `navidrome` service only used `expose` (container-network-only).
   Switched to `ports: - "127.0.0.1:4533:4533"` — loopback-only, so Navidrome stays exactly
   as unreachable from the internet as before (still only `cloudflared` reaches it
   externally), while becoming reachable from the box itself.
5. **Same call's error handling was also incomplete**: the login `curl` had no failure
   guard, so `set -e`/`pipefail` killed the *entire* import run on any Navidrome-unreachable
   condition, not just the graceful "skip the scan trigger" path the function's comments
   already described for a failed login. Added `|| true` to match.
6. **The native admin scan-trigger endpoint doesn't exist in the pinned Navidrome
   `0.63.2`**: confirmed against upstream source (`server/subsonic/library_scanning.go`,
   `server/subsonic/api.go` line ~181) — `StartScan` is registered as the *Subsonic* API's
   admin-only `startScan` action (query-param auth: `u`/`p`/`v`/`c`), not a native
   `/api/scanner/scan` JWT-bearer endpoint. Every `music-import` run now logs a 404 there
   after a real login success. **Not fixed this pass** — left as a known, non-fatal bug
   (Navidrome's own `ND_SCANNER_SCHEDULE: "0 3 * * *"` cron backstop still keeps the library
   current) rather than rush an unverified Subsonic-auth rewrite. Follow-up: rewrite
   `trigger_navidrome_scan()` to call `GET /rest/startScan` with Subsonic auth params.

**Two gaps could not be closed this pass, both blocked by the Claude Code auto-mode
classifier refusing an on-box `dnf install` from a subagent** (same shape as Task 5's
DELETE/create-tunnel refusal above — mutating infra calls need Geoff present, not a
subagent acting alone):

- **`ffmpeg`/`flac` are still absent on this box.** The EPEL + RPM Fusion `runcmd` block
  that installs them (`cloud-init/user-data.yaml`) was added to the repo *after* this box
  was provisioned (Task 4a), so it never ran here — a cloud-init-only fix never converges
  an already-running box. beets' `replaygain` plugin fails to load on every `music-import`
  invocation as a result (logged, non-fatal against an empty inbox; will break real
  imports needing the quality gate's `flac -t`/`metaflac` and replaygain analysis).
- **`music-backup.service` cannot run at all**: `sqlite3` and `rclone` are both missing on
  the box, and neither was ever listed in `cloud-init/user-data.yaml` in the first place (a
  real spec gap, not just an unapplied one). Added `sqlite` to the `packages:` list (base
  AlmaLinux repo, confirmed by lookup) for future re-provisions; `rclone` and `msmtp` (also
  missing, also never listed — `lib.sh`'s `notify()` needs it, untested this pass since an
  empty-inbox run never reaches it) are left as an open item rather than guessed into an
  unattended `set -e` first-boot script without a confirmed AlmaLinux 10 install path.
  Confirmed via a read-only `rclone lsjson` against `music-library` that `state/` is empty,
  consistent with no backup ever having run.

Acceptance sweep otherwise green: all three containers running with zero restarts after a
60s settle; the tunnel reads `status: healthy` with active connections; `music.907.life`
serves Navidrome's web UI with no Access interstitial and answers `/rest/ping`;
`inbox.907.life` presents the Cloudflare Access challenge before any FileBrowser content;
both timers active; the Navidrome admin account was created via `POST /auth/createAdmin`
and its login verified via `POST /auth/login` (API only, no secret values ever printed —
checked by name, length, and boolean fields throughout this whole pass).

## Task 5: Cloudflare tunnel wiring, plus a same-day secret rotation (2026-08-31)

Geoff added Cloudflare Tunnel:Edit to `CLOUDFLARE_API_TOKEN` 2026-08-30, unblocking Task 5.
An agent created tunnel `musicbox` (`POST .../cfd_tunnel`), then displayed that call's full
JSON response in a transcript while reporting progress. **Wrong to rediscover**: a
`cfd_tunnel` CREATE response body carries the tunnel's `TunnelSecret` and its connector
`token` alongside the id — displaying or logging the raw body exposes both, not just an
identifier. The fix at the call site is mechanical: always pipe the response through
`jq -r '.result.id'` (or whichever single field is actually needed) and never let the full
body reach a transcript, a log, or a report.

Once flagged, the tunnel was treated as compromised under standing incident doctrine (a
secret printed into a transcript is compromised, full stop — no risk-assessment exception).
Remediation was delete-and-recreate, the cheap definitive path since `cloudflared` was not
yet deployed anywhere against the exposed token:

1. Deleted tunnel `66fa450c-5abb-459d-a1f8-b80d8d19e07c` (`DELETE .../cfd_tunnel/{id}`).
2. Created a new tunnel under the same name `musicbox`, this time piping the response
   through `jq -r '.result.id'` only: id `b6781175-000b-40c3-a54a-05906ab25037`.
3. Fetched its token (`GET .../cfd_tunnel/{id}/token`) and piped it directly into
   `secret-set.sh MUSICBOX_TUNNEL_TOKEN --stdin`, overwriting the old value. Verified
   name-only (`grep -c` on `~/.local/secrets`, count 1) and by length only (240 chars) —
   never printed.
4. Re-PUT the ingress from `config/cloudflared-tunnel-config.json` to the new tunnel id and
   verified with GET: `music.907.life` → `http://navidrome:4533`, `inbox.907.life` →
   `http://filebrowser:80`, 404 catch-all.
5. Repointed both DNS CNAMEs (`music.907.life`, `inbox.907.life`, zone `907.life`) to
   `b6781175-000b-40c3-a54a-05906ab25037.cfargotunnel.com`, proxied.

**Process note**: the DELETE and the tunnel-recreate POST were both refused by the Claude
Code auto-mode classifier when a subagent attempted them (destructive/mutating infra calls
without the user present); read-only GETs against the same account worked fine throughout.
The rotation completed once the conductor ran those two calls from the main session with
Geoff present, then handed the new tunnel id back for the remaining wiring and docs.

Cloudflare Access was untouched throughout (confirmed read-only via `CF_ZT_TOKEN`:
`inbox.907.life` is the only Access app on the musicbox surface; `music.907.life` stays
Access-free). `cloudflared` is not deployed on the box yet, so both hostnames 530 until
Task 4's first real `deploy.sh` — expected, not chased.

## Task 2 follow-up: quality gate, ping hardening, config sync (2026-08-30)

TDD, red-first for every case, against the spec's freshly amended import-pipeline
section (step 4). Four items, one commit each:

- **`quality_gate <dir>`** (`scripts/lib.sh`): walks every audio file under dir
  (case-insensitive flac/mp3/m4a/ogg/opus/wav/aiff extensions; anything else is out
  of scope). A non-FLAC extension fails outright; a FLAC must pass `flac -t`'s full
  decode test and carry >=44.1kHz/16-bit (`metaflac --show-sample-rate
  --show-bps`). Prints "path\treason" per failure, exit 0 only if everything
  passed. Wired into `music-import` as `quarantine_quality_failures`, between the
  free-space abort and `beet import`: any failing file takes its **whole**
  containing album directory to `review/<date>/<user>/` (a partial import of just
  the good tracks never happens), never just the offending file.
- **Uploader copy is templated, not generated**: `config/reject-email.txt` is
  Geoff's verbatim consumer-facing copy with three literal placeholders (`{first
  name}`, `{album or file name}`, `{one reason line per file, from the mapping
  below}`), filled by plain bash substring replacement in
  `render_reject_email()` — no sed, no metacharacter risk. `friendly_reason()`
  translates `quality_gate`'s raw technical reason into one of three fixed human
  sentences (non-FLAC extension, failed decode, below CD quality); Geoff's own
  copy of the notice stays technical (raw path + reason), same asymmetry as the
  existing beets-skip notice. First name comes from `uploader_name()`, the first
  word of the matched `contributors.yaml` `name` field (mirrors `uploader_email`'s
  lookup). Default template path is `/opt/musicbox/config/reject-email.txt` —
  `config/` is already rsynced wholesale to that tree by `deploy.sh`, so this
  needed no new deploy step; only `contributors.yaml` and `beets-config.yaml` get
  dedicated ones, for permissions/install-location reasons this file doesn't
  share.
- **`ping_hc` hardened**: `--retry 3 --max-time 10` added to its `curl` call, so a
  slow or flaky healthchecks.io can never be the reason an otherwise-successful
  run hangs or gets misreported. The empty-vs-unset URL guard already existed;
  only a new bats case (env var set but empty) was added to pin it down.
- **Flock-held early exit now pings success before `exit 0`** (spec decision):
  a run that finds the lock already held (another scheduled run, or an
  interactive `music-beet` session) is the healthy case this check exists for,
  not a failure — healthchecks.io has no third state between success and
  failure. Accepted residual: a genuinely wedged import would also report
  healthy via this path until someone notices stale music in the library.
- **Beets config synced to the workstation's** (`~/.dotfiles/beets/.config/beets/config.yaml`):
  `original_date: yes`, `per_disc_numbering: yes`, and the `replaygain` plugin
  (backend `ffmpeg`), added to `config/beets-config.yaml` alongside the existing
  quiet-mode section. flac (the quality gate) and ffmpeg (replaygain's analysis
  backend) are installed on the host via a dedicated `cloud-init/user-data.yaml`
  runcmd step, not the top-level `packages:` list (that module runs before
  runcmd, too early to enable a repo first): `dnf -y install epel-release`,
  `dnf config-manager --set-enabled crb` (both EPEL and RPM Fusion need it),
  then the RPM Fusion free EL10 GPG key is fetched and `rpm --import`ed and
  its release package installed against it (gpgcheck on throughout, no
  `--nogpgcheck`), then `dnf install -y flac ffmpeg`. Ruling: both binaries
  parse untrusted contributor uploads, so they stay on dnf-automatic's
  security-update path rather than a pinned static build, which would freeze
  a codec-parsing attack surface instead of patching it.
- **Known limitation, accepted**: `quality_gate`'s output channel is
  tab/newline-delimited ("path\treason" per line), so a filename containing a
  literal tab or newline would misparse downstream in
  `quarantine_quality_failures`. The `find -type f -print0` walk itself is
  correct (NUL-delimited, handles any filename); the limitation is only in
  what quality_gate prints afterward. Not expected in practice for
  contributor uploads; revisit if it ever is.

Fixed a docs nit found while touching STATUS.md: Task 4a's on-box checks table
(above, Task 4a's own section) lists nine checks but was labeled "8 of 8"; now "9
of 9" in both STATUS.md and here.

Full test suite: `tests/lib.bats` (4 new `quality_gate` cases + 1 `ping_hc` case)
and `tests/music-import.bats` (1 new flock-ping case, 1 new quality-gate
end-to-end case; three existing fixtures switched from placeholder `.flac` text to
real ffmpeg+flac-encoded audio, since the gate now actually decodes them).
`bash scripts/check.sh` exits 0.

## Task 2 follow-up, fix round: Opus review of 5da3844..d43972b (2026-08-30)

Six items landed, TDD where a test was named:

- **Blocking: unquoted `&` in `render_reject_email`'s `${var//pat/repl}`
  substitutions.** Confirmed on this box's bash (5.3.9): an unquoted
  replacement containing a literal `&` gets expanded to the matched pattern
  text, exactly like sed — an album named "Sturm & Drang" rendered as "Sturm
  {album or file name} Drang" instead of the real name. Fixed by quoting
  every replacement value; bats case covers the ampersand literally.
- **Blocking: an `mv` into an already-occupied `review/<date>/<user>/<album>`
  aborted the whole run under `set -e`** (a same-day repeat quarantine of an
  identically named album is the common case, not an edge case). New
  `lib.sh` helpers: `unique_dest_path` (appends a `-HHMMSS` suffix on
  collision, then `.2`/`.3`/... if even that collides) and
  `safe_mv_to_review` (uses it, and warns-and-returns-nonzero instead of
  letting a failed move take the run down). Applied to both
  `quarantine_quality_failures` and `move_skips_and_notify`, which shared the
  same unguarded-`mv` shape. bats case: two identically named "Bad Album"
  quarantines in one test run both survive and beet import still runs both
  times.
- **A missing `reject-email.txt` template previously sent the uploader an
  empty-bodied email** (`render_reject_email`'s failure was silently
  swallowed by command substitution). Now an explicit `[ -f ... ]` check
  before the `notify` call skips the uploader-facing email and logs why;
  Geoff's technical summary still covers the album either way.
- **Each quality-gate reason line is now prefixed with its file's
  basename** in the uploader-facing email, so an album with more than one
  failing file reads as N per-file lines instead of running together.
- **Added the 8-bit-FLAC branch's missing test coverage** (`ffmpeg -acodec
  pcm_u8` piped into `flac`), the one `quality_gate` branch that had no
  case before this round.
- **CRB enablement** added to cloud-init's EPEL/RPM Fusion runcmd block
  (`dnf config-manager --set-enabled crb`, right after `epel-release`):
  both repos need packages CRB carries, and ffmpeg's dependency chain fails
  without it.

Follow-up verification round found one more leak in the `mv`-collision fix:
`safe_mv_to_review` failing for a **non-collision** reason (permissions, a
bad mount) still `continue`d out of the loop with no Geoff notice and the
album still sitting in staging — exactly what `beet import`'s unconditional
recursive scan of `$STAGING_DIR` would then ingest, silently defeating the
whole quality gate. Fixed with a second `safe_mv_to_review` attempt to
`$MUSICBOX_ROOT/quarantine-failed/<date>/<user>`, a holding dir outside both
staging and review's own tree (so a problem isolated to `review_root`
doesn't also block the fallback); Geoff's summary gets a `could not move;
excluded from import` marker either way, and a distinct all-caps marker for
the (untested, believed-unreachable without a fully read-only filesystem)
case where even the fallback fails. The holding dir's root is a top-level
constant (`QUARANTINE_FAILED_DIR`) beside `REVIEW_DIR` and the rest of the
tree, deliberately not created up front: a healthy box never has one.

Also fixed in `safe_mv_to_review` itself: its `mkdir -p "$dest_dir"` call
was unguarded, so a destination that couldn't even be created would have
aborted the whole run under `set -e` — the exact failure mode the function
exists to prevent. Added a matching
`[ -n "$geoff_summary" ] || return 0` guard before the final Geoff `notify`,
defense-in-depth alongside the pre-loop `album_order` check (every branch of
the loop already appends something, so this is currently unreachable, but
decouples "should Geoff be emailed" from "was there at least one album_key"
for whatever a future change to the loop might do). TDD: bats case forces
the primary move to fail via `chmod 0555` on the review dir and asserts the
album lands in `quarantine-failed`, never in staging, with the marker in
Geoff's email and no uploader email.

Also fixed, while adding coverage for the above: three existing
`tests/music-import.bats` assertions used `! grep ...` as a non-final line —
under bats' `set -e`, a command negated with a leading `!` never triggers
errexit regardless of its actual result, so a failing negated assertion that
isn't the test's last line silently does nothing. Converted to bats' `run` +
`[ "$status" -ne 0 ]` idiom throughout the file for safety, whether or not a
given instance was actually latent.

## Task 5/6: Cloudflare Access + Navidrome secrets, tunnel blocked (2026-08-30)

Control-plane pass, no box changes. Findings and what landed:

- **Tunnel creation blocked, not a code problem**: `POST
  /accounts/{a}/cfd_tunnel` 403s `{"code":10000,"message":"Authentication
  error"}` on both `CLOUDFLARE_API_TOKEN` and `CF_ZT_TOKEN` — neither carries the
  Cloudflare Tunnel:Edit permission group (confirmed: both tokens list/read
  tunnels fine, i.e. GET `.../cfd_tunnel` is 200; only the POST 403s). This
  blocks Task 5's steps 1-4 (create tunnel, fetch token, PUT ingress config, DNS
  CNAMEs to `<tunnel-id>.cfargotunnel.com`) and, downstream, Task 4's first real
  `deploy.sh` run (cloudflared needs `MUSICBOX_TUNNEL_TOKEN` to come up).
  Resolution needs Geoff to add Cloudflare Tunnel:Edit to one token via the
  dashboard (API self-management of token scopes is deliberately locked out).
  Detail in registry.md's `CF_ZT_TOKEN` entry.
- **Cloudflare Access, done**: confirmed no Access application exists on any
  `907.life` hostname (checked the full account app list — zero matches), so
  `music.907.life` correctly has none. Created a self-hosted Access application
  for `inbox.907.life` (`CF_ZT_TOKEN`, which — unlike `CLOUDFLARE_API_TOKEN` —
  can read Access identity providers) restricted to the account's existing
  `onetimepin` identity provider (`e33c7c2d-facd-446c-96de-8215f64cde76`,
  `auto_redirect_to_identity: true` since it's the only allowed IdP), with an
  allow policy for `geoff-login@907.life` (policy
  `92425260-87f6-4a3a-b66e-d71bb8d9d6a8`). Group emails join this policy at
  Task 8 per plan. **Not yet verified end to end**: the "Access challenge
  appears before FileBrowser's login" acceptance check needs the tunnel serving
  traffic first, so it waits on the tunnel-permission fix above.
- **Task 6 Navidrome secrets, done**: generated and stored
  `ND_PASSWORDENCRYPTIONKEY` (openssl rand -hex 32), `NAVIDROME_ADMIN_PASS`
  (openssl rand -hex 16), `NAVIDROME_ADMIN_USER` (`admin`) via `secret-set.sh`.
  No value printed; lengths only (64/32/5 chars) confirmed against
  `~/.local/secrets`.
- **registry.md, done**: added rows + per-secret sections for
  `MUSICBOX_TUNNEL_TOKEN` (marked not-yet-minted with the blocker noted),
  `ND_PASSWORDENCRYPTIONKEY`, `NAVIDROME_ADMIN_USER/PASS`, and the four R2 pairs
  already in the store (`MUSICBOX_R2_*` — musicbox-vps token, Object Read &
  Write, bucket `music-library`; `MUSIC_R2_RO_*` — claude-code-thinkpad-x1
  token, Object Read only, same bucket); noted `HCLOUD_TOKEN`'s 2026-08-30
  rotation. Committed to `~/.dotfiles` as two commits (registry.md text, then
  `values.age` separately) — a third file in the working tree,
  `bin/.local/bin/secret-receive`, was mid-edit from an unrelated session and
  deliberately left untouched/uncommitted.

Decision a later pass would be wrong to rediscover from scratch: **use
`CF_ZT_TOKEN`, not `CLOUDFLARE_API_TOKEN`, for anything touching Access
identity providers** — the general-purpose token gets `auth.forbidden` reading
`.../access/identity_providers`, even though it can read/write Access apps and
policies fine.

## Task 4a: AlmaLinux 10 image-defect fix round, rebuild to 8/8 green (2026-08-31)

Two deterministic image defects blocked the recreated box at 5/8 green (see the prior
round's entry below): Docker's missing `xt_addrtype` kernel module
([docker/docker-ce-packaging#1340](https://github.com/docker/docker-ce-packaging/issues/1340))
and the Hetzner `alma-10` image shipping `SELINUX=permissive` instead of the distro
default of enforcing. Both fixes landed in `cloud-init/user-data.yaml`'s `runcmd` plus a
new top-level `power_state` key:

- **Docker/kernel-module fix**: a new `runcmd` item installs
  `kernel-modules-extra-$(uname -r)` — the exact NVR paired to the boot's running
  kernel — before the Docker CE install, falling back to `dnf update kernel &&  dnf
  install kernel-modules-extra` only if the mirrors have moved past this image's shipped
  kernel (that fallback needs the reboot below to take effect). `systemctl enable --now
  docker` became `systemctl enable docker` — starting it before the reboot would still run
  under the old kernel and fail identically.
- **SELinux fix**: `sed` flips `/etc/selinux/config` to `SELINUX=enforcing` and `touch
  /.autorelabel` schedules a full relabel on next boot (the safe path — `setenforce 1`
  alone would trust every file on a vendor image already carries correct labels, which
  permissive mode never enforced).
- **Reboot ordering**: a single `power_state: {mode: reboot}` module — not a `reboot` call
  inside `runcmd` — runs after every other cloud-init module completes and has its
  once-per-instance completion recorded, covering both fixes in one reboot. The
  autorelabel triggers its own second, self-contained reboot; the acceptance run must
  tolerate SSH dropping twice.

First attempt at this round left the fix uncommitted while re-verifying against the live
box — caught by a review pass that diffed the deployed `/var/lib/cloud/instance/user-data.txt`
sha256 against `HEAD`'s copy of `cloud-init/user-data.yaml` and found them different
(`196e76f...` deployed vs `db0b633...` at HEAD). Lesson for any future fix-round: verifying
green against a live box proves nothing about the repo until the fix is committed and that
commit's own blob hash is re-checked against the deployed file — a worktree edit and an
unpushed commit look identical to `ssh`.

Re-verified after commit `3fa905d`: `sha256sum cloud-init/user-data.yaml` at HEAD equals
the deployed `/var/lib/cloud/instance/user-data.txt` hash
(`196e76fbc51d5a5828c6a378761d480b790160cc9fd07fa7a1f41ab1557dfa5a`). `bash
scripts/check.sh` exits 0.

**9 of 9 on-box checks green, re-verified 2026-08-31 against `root@5.78.148.42`:**

| Check | Command | Result |
|---|---|---|
| cloud-init | `cloud-init status --wait` | `status: done` — PASS |
| volume mount | `mountpoint -q /srv/music` | mounted — PASS |
| tree ownership | `ls -lnd /srv/music/{library,inbox,review,state,staging}` | all `2000 2000`; `inbox` is `drwxrwsr-x` (2775, setgid) — PASS |
| music user | `id music` | `uid=2000(music) gid=2000(music) groups=2000(music)` — PASS |
| docker | `systemctl is-active docker` | `active` — PASS |
| SELinux | `getenforce` | `Enforcing` — PASS |
| sshd password auth | `sshd -T \| grep -i '^passwordauthentication'` | `passwordauthentication no` — PASS |
| yq / jq | `/usr/local/bin/yq --version` / `jq --version` | `v4.53.6` / present (`jq-` is a cosmetic upstream version-string quirk on this AlmaLinux 10 build, not a functional defect) — PASS |
| fstab | `grep srv/music /etc/fstab` | exactly one line: `UUID=6614e3dd-9014-4949-9f4c-effcd317bd3a /srv/music ext4 defaults,nofail,x-systemd.device-timeout=30s 0 2` — PASS |

## Task 4a: real provisioning, forced-root-password fix round (2026-08-30/31)

`scripts/provision.sh --check` matched expectations (alma-10 available, no fallback
needed), then `scripts/provision.sh` created firewall `musicbox-fw` (tcp/22 only, both
address families), volume `musicbox-data` (50GB, `hil`), and server `musicbox` (CPX21,
`alma-10`, `hil`). Server IPv4 `5.78.148.42`, IPv6 `2a01:4ff:1f0:97c6::1`. CPX21 traffic
allowance at `hil`: 2199023255552 bytes = 2 TiB/month (`hcloud server-type describe cpx21
-o json`), overage $1.20/TB — generous for a group Navidrome, no need to default clients
to transcoded streams.

A fix round closed three blocking findings from a forced-root-password blocker, and the
server was recreated to pick them up:

1. `scripts/provision.sh` registers an `hcloud ssh-key` named `musicbox-workstation`
   idempotently and passes `--ssh-key musicbox-workstation` to `hcloud server create`.
   Confirmed on the recreate: no "Root password: ..." line appeared in the create output.
2. `cloud-init/user-data.yaml`'s `runcmd` carries a defense-in-depth guard — `chage -M -1
   -I -1 -E -1 -d "$(date +%Y-%m-%d)" root` — as the first item, so any future path that
   creates the server without a key can't reintroduce this blocker. Confirmed on-box:
   `chage -l root` shows `Password expires: never`.
3. The prior round's one-time-manual-SSH-unblock instruction was removed; nothing was
   hand-applied to a live box — the box was deleted and recreated from the fixed repo
   instead (volume `musicbox-data` survived and reattached, `blkid` guard prevented
   reformat, box got IPv4 `5.78.148.42` again).

Re-verification after that round found a new, unrelated blocker: `cloud-init status
--wait` returned `status: error` because Docker failed to start
(`docker/docker-ce-packaging#1340`, see the entry above), which also aborted the
downstream yq-install `runcmd` item. `getenforce` also reported `Permissive`, traced to
the Hetzner `alma-10` image's own shipped default, not to anything cloud-init touched.
Both became the next fix round, entry above.

## Task 3: service configuration — compose, systemd units, deploy (2026-08-30)

Landed `compose.yaml` (Navidrome, FileBrowser Quantum, cloudflared — no service
publishes a host port; only cloudflared is reachable from outside), `config/
navidrome.toml`, `config/filebrowser.yaml`, `config/msmtprc`, `config/beets-config.yaml`,
`config/cloudflared-tunnel-config.json`, the four `systemd/music-{import,backup}.
{service,timer}` units plus a shared `systemd/musicbox-ping-failure@.service` template,
`env/musicbox.env.template`, and `scripts/deploy.sh`. Extended `scripts/check.sh` with
`docker compose config -q` and a pattern-filtered `systemd-analyze verify`. All pinned
tags looked up and dated in `docs/STATUS.md`.

Two Task 2/Task 3 interface mismatches found and fixed (Task 3's own acceptance criteria
named the authoritative env var scheme, so Task 2's code was the thing that moved):
- **healthchecks.io slugs**: `music-import`/`music-backup` called `ping_hc
  musicbox-import` / `ping_hc musicbox-backup`, which resolves (per `lib.sh`'s own
  uppercase-and-dash-to-underscore rule) to `MUSICBOX_HC_MUSICBOX_IMPORT_URL` /
  `MUSICBOX_HC_MUSICBOX_BACKUP_URL` — doubled "MUSICBOX", not the
  `MUSICBOX_HC_IMPORT_URL` / `MUSICBOX_HC_BACKUP_URL` this task's env template requires.
  Fixed both scripts to call `ping_hc import` / `ping_hc backup`; updated
  `tests/lib.bats` and `tests/music-import.bats` to match (still 16 green). Monitoring
  would otherwise have silently no-op'd forever — `ping_hc` treats an unset URL var as a
  no-op, not an error, precisely so a monitoring bug never takes down the pipeline it
  monitors, which also means a naming bug like this one would never have surfaced as a
  failure anywhere.
- **Navidrome scan-trigger credentials**: `music-import`'s `trigger_navidrome_scan()`
  read `NAVIDROME_USER`/`NAVIDROME_PASSWORD`; this task's env template is required to
  name `NAVIDROME_ADMIN_USER`/`NAVIDROME_ADMIN_PASSWORD`. Renamed in the script and in
  `tests/music-import.bats`.

Two review findings fixed after the first pass (both would have broken the first real
deploy, not just this repo's own gate):
- **`/etc/musicbox/env`'s parent directory didn't exist.** Nothing in `provision.sh` or
  `cloud-init/user-data.yaml` creates `/etc/musicbox` (only `/srv/music` is built at
  provision time), so `install -m 600 ... /etc/musicbox/env` died with "cannot create
  regular file" on a fresh box, after the scripts and units had already landed. Fixed by
  adding `-D` to that `install` call, which creates the missing parent and is a no-op on
  every later re-deploy.
- **`config/msmtprc` was never actually installed to `/etc/msmtprc`.** The file's own
  header already claimed deploy.sh put it there, but deploy.sh only rsynced `config/` to
  `/opt/musicbox/config` for the record — `lib.sh`'s `notify()` calls plain `msmtp`, which
  resolves its account from `/etc/msmtprc` and would have found nothing there, aborting
  `music-import` under `set -euo pipefail` on the first skipped-track notification. Added
  an explicit install step (rsync to a temp path, then `install -m 0644` to
  `/etc/msmtprc`), with a matching diff-and-report line in `--check`.

Decisions a later pass would be wrong to rediscover from scratch:
- **Navidrome config split**: non-secret baseline settings (`MusicFolder`, `DataFolder`,
  `LogLevel`) live in `config/navidrome.toml`, bind-mounted read-only at
  `/data/navidrome.toml` alongside the read-write `/data` mount for the DB — Docker lets
  a single-file bind mount overlay a path inside an already-mounted directory. Scanner
  settings and the encryption key live in `compose.yaml`'s `environment:` instead (ND_
  env vars override the toml), specifically so the two option names the task called out
  by name (`ND_SCANNER_SCHEDULE`, `ND_SCANNER_WATCHERWAIT`) are literally grep-able in
  compose.yaml. `ND_CONFIGFILE=/data/navidrome.toml` is set explicitly rather than relied
  on implicitly, per navidrome.org's own Docker install guidance.
- **FileBrowser Quantum 1.5.x has no static per-user YAML config** — verified against the
  actual `v1.5.5-stable` tag's Go source (`backend/common/settings/structs.go`), not just
  the docs site (which mixes in v2.0 beta content). Users and their scopes live in its own
  SQLite DB, created via CLI/API, not this repo. `config/filebrowser.yaml` instead
  configures one source (`/srv/music/inbox`) with `createUserDir: true` +
  `defaultUserScope: "/"`, so creating an account named `alice` automatically scopes her
  to exactly `/srv/music/inbox/alice` — no per-user YAML needed, and the username IS the
  inbox directory name music-import already matches against `contributors.yaml`. Task 8
  creates the real accounts (one per `contributor`-role entry) once Task 0 lands real
  names.
- **cloudflared runs `tunnel --no-autoupdate run` with no local `--config`.** A
  token-driven tunnel is remotely managed; a local ingress file is silently ignored
  (verified against upstream `cloudflared` issues, not assumed). The actual ingress
  mapping is `config/cloudflared-tunnel-config.json`, a ready-to-PUT API body for Task 5
  to apply once the named tunnel exists — see STATUS.md's Open items, this is a real gap
  Task 5 must not skip.
- **msmtp's `passwordeval` reads `$MUSICBOX_SMTP_PASSWORD` from its own inherited process
  environment** (`printenv MUSICBOX_SMTP_PASSWORD`), not by re-reading
  `/etc/musicbox/env` from disk. That variable is already present in music-import's
  environment via its unit's `EnvironmentFile=`; having msmtp open the file a second time
  would require it group-readable by the unprivileged `music` user, widening every OTHER
  secret in that one file to the same process for no actual reduction in blast radius
  (that process already receives all of them via `EnvironmentFile=`). `config/msmtprc`
  itself holds no secret and installs world-readable at `/etc/msmtprc`.
- **Deploy target directory**: `deploy.sh` rsyncs `compose.yaml` + `config/` to
  `/opt/musicbox` on the box (not `/srv/music`, which is data, restored from R2, and
  never holds repo-sourced config). `docker compose` runs with `--project-directory
  /opt/musicbox` so the compose file's relative `./config/...` volume paths resolve
  regardless of the caller's cwd.
- **`deploy.sh --check` never prints secret contents**, even when diffing the rendered
  `/etc/musicbox/env` against what's already on the box: it reports only whether the
  content differs (`diff -q` via process substitution, nothing touches local disk).
  Exercised in this pass against an intentionally-unreachable host and against
  `127.0.0.1` (connection refused) — both paths report cleanly and exit 0 without a real
  box to deploy to.
- **beets on the box**: installed via `uv tool install --force beets==2.13.1` as the
  `music` user (`HOME=/srv/music`), then symlinked `/usr/local/bin/beet ->
  /srv/music/.local/bin/beet` — `uv tool install`'s default bin dir is not on systemd's
  default `PATH`, so without the symlink `music-import`'s `beet import` call would fail
  with "command not found" the first time a real timer fires. `config/beets-config.yaml`
  mirrors the workstation's `~/.dotfiles/beets` layout and plugins (fetchart, embedart),
  repointed at `/srv/music`, plus a quiet-mode section (`quiet_fallback: skip`,
  `none_rec_action: skip`) so `beet import -q` sends anything it can't confidently tag to
  review instead of guessing or blocking on input that will never arrive.

## Task 2: import, backup, and wrapper scripts (2026-08-30)

Landed `scripts/lib.sh` (`quiescent`, `safe_extract`, `free_space_ok`, `notify`,
`ping_hc`), TDD with `tests/lib.bats` written and confirmed red before any
implementation existed. Landed `music-import`, `music-backup`, `music-beet`, and
`scripts/check.sh` as the repo gate (shellcheck over every `scripts/*` bash file,
then the full bats suite). `tests/music-import.bats` runs the importer end to end
against a fixture `/srv/music` tree with `beet`, `curl`, and `msmtp` stubbed on
PATH: happy path, a beets skip filed to review with both notifications, a
non-quiescent inbox left untouched, and a free-space abort that fires the fail
ping before beet ever runs.

Two one-line review cleanups from Task 1 landed alongside: the volume glob guard
in `cloud-init/user-data.yaml` now also rejects a non-block-device match (a
zero-match glob previously fell through as a false positive on some shells), and
`provision.sh`'s firewall-rule check now fails with its own message when the
`hcloud firewall describe` call itself errors, instead of reporting a network
blip as a missing rule.

Decisions a later pass would be wrong to rediscover from scratch:
- **Contributor matching**: `music-import` matches an inbox directory name
  (`/srv/music/inbox/<user>`) against contributors.yaml's `name` field verbatim;
  there is no separate username/slug field. If Task 3/8's FileBrowser account
  provisioning wants a different inbox directory naming scheme, `contributors.yaml`
  needs a dedicated field and `uploader_email()` needs updating to match on it.
- **Navidrome scan trigger**: implemented against the native admin API (`POST
  /auth/login` for a JWT, then `POST /api/scanner/scan` with `x-nd-authorization:
  Bearer <token>`), not the Subsonic API, because Subsonic has no scan-trigger
  endpoint. This is the UI's internal API rather than a documented stable
  contract; verify it against the pinned Navidrome image tag once Task 3 chooses
  one, since a trigger failure is only best-effort (logged, not fatal) precisely
  because this endpoint is not guaranteed stable.
- **State DB paths default to** `$MUSICBOX_ROOT/state/library.db` (beets) and
  `$MUSICBOX_ROOT/state/navidrome/navidrome.db` (Navidrome), both overridable via
  env var. Task 3's `compose.yaml` must either mount Navidrome's data dir there or
  the env var overridden in `/etc/musicbox/env`.
- **Archive cap default**: `MUSICBOX_MAX_ARCHIVE_BYTES` defaults to 2 GiB in
  `music-import`; revisit if a real Bandcamp download exceeds it.
- New host dependency: `yq` (mikefarah/go, v4) for parsing `contributors.yaml`,
  needed on the box alongside `beet`/`rclone`/`msmtp`/`bsdtar`. Add it to Task 3's
  `deploy.sh` package list.
