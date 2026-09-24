# Draft HISTORY entry: rung 5, pass 5b-ii (auth handlers, mail seam, limiters, close)

**Draft only. The conductor folds this into `docs/HISTORY.md` at the merge, path-limited.**

---

## Rung 5, pass 5b-ii (the roster, the mail seam, the rate limiters, the origin check, the auth handlers, the security fold): closed 2026-09-23, lane 1

Executed against `docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md`
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
`cairn-run-gate` at this close; the wrapper does not print a wall-clock
total, so no number is claimed for it beyond "green".

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
channel, address, admin flag, and client IP. The admin flag may be set
only when the caller's channel is `identity.ChannelServiceToken`,
refused on every other channel by its own sentinel — the conductor's
ruling that the invariant lives in the write path, not in 5c's caller.
The fix round moved the admin-channel refusal from "every channel but
one" partial coverage to every non-service-token channel, and gave
`doctor` a row that reports a missing roster as a failure when a
selected source requires one.

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

### The twelve resolutions (R-5b17 through R-5b28)

Each carries its marker's site and what execution confirmed. **Only one
site uses the plan's literal marker text** (`// ruled, conductor
2026-09-12, R-5b2, R-5b3` at `cmd/dubplate-server/auth.go:41`, which is
a 5b-i-era marker, not one of these twelve); the twelve below cite their
resolution number in a plain parenthetical (`(R-5bN)`) rather than the
"ruled, conductor" stamp the plan's own text at `:3372-3378` describes.
**Corrected below**, since the grep the plan names does not recover
them.

| Resolution | Decided | Landed | Confirmed by execution |
| --- | --- | --- | --- |
| R-5b17 | 5b-i's security fold is Tasks 1-2, not a separate pass | Tasks 1-2 | Confirmed: B1/M1/M2/M3/M5 fixed before Task 3 dispatched |
| R-5b18 | 5b-i's security read is discharged; this pass's own read covers both halves at close | pass-end review, `46e37d1` | Confirmed: the review's scope line covers `1d4ed29..HEAD`, both halves |
| R-5b19 | roster file is `{"schema_version": 1, "members": [...]}`, not a bare array | `internal/auth/roster.go:19,98` | Landed as decided |
| R-5b20 | the store creates the file, never the directory; pre-create is 5d's | `internal/auth/roster.go:140` (`CreateRoster`) | Landed as decided; `config/roster.example.json` created here rather than in 5d (recorded, spec correction below) |
| R-5b21 | the audit record is a structured `slog` record; no audit file opens | `internal/auth/roster.go` (`AddMember`, Task 5) | Landed as decided |
| R-5b22 | the unknown-address answer is `KindString`, default `kind`, values `kind`/`generic` | `cmd/dubplate-server/auth.go:197-280` | Landed as decided |
| R-5b23 | the limiter map bound and sweep period are package constants, not thresholds | `internal/auth/limit.go`, `internal/auth/nonce.go` | Landed as decided |
| R-5b24 | three limiters, three key spaces, not one shared map | `internal/auth/limit.go` (`keyedLimiter`, three instances) | Landed as decided; **the security review's finding 3 shows the per-address key space is still charged before the per-client check refuses, an ordering gap R-5b24 did not itself cover** |
| R-5b25 | handlers are `http.Handler` values from one dependency struct, no package-level route registration | `internal/auth/handlers.go` | Landed as decided |
| R-5b26 | no redirect-target validator; R33 travels as a constraint to rung 12 | disposed, no function added | Confirmed: Task 0 item 10 and this close both grep `internal/auth` and find none; the `b91bae1` erratum's contrary claim is corrected below |
| R-5b27 | the origin check inserts between normalization and prefix routing | `internal/serve/serve.go:159-172` (`Handler`) | Confirmed by reading `Handler`'s wrap order: strip, Normalize, refuseEchoMethods, origin check, routed |
| R-5b28 | the generic answer's ordering: `kind` governs R9, `generic` answers before the send | `internal/auth/handlers.go` (Task 10) | Confirmed by the security review: "the generic answer is identical in bytes and in measured timing" |

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
   in 5c's caller. Confirmed: `AddMember` (`internal/auth/roster.go`,
   Task 5) refuses `Admin` on every channel but
   `identity.ChannelServiceToken`, and the security review's
   "attacked and found sound" list confirms the seam reads `Admin` only
   from the roster, never the request.
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

`docs/superpowers/research/2026-09-23-rung-5b-ii-security-review.md`:
scope `git diff 1d4ed29..HEAD -- '*.go' go.mod go.sum config/'` at
`08befea`. **No blocker.** Five majors (cross-process roster writes,
mailed links landing outside the owned prefix, a refused client
evicting a victim's per-address bucket, IPv6 per-client keys at /128,
no trusted-proxy setting) and fourteen minors.

**Disposition of the twenty 5b-i findings this review re-checked**
(the do-not-relitigate baseline for 5c and 5d):

| Finding | Status | Evidence |
| --- | --- | --- |
| B1 | Fixed | 1s retention, 48h token; at +25h, Removed 0, replay refused |
| M1 | Fixed | A quoted space or comma is refused; every accepted corpus entry is a fixed point |
| M2 | Fixed within one process | A backward bump returns `ErrEpochBackward`; cross-process is finding 14 below |
| M3 | Fixed | 0 of 64x6 final-character variants verify |
| M4 | Partial | `doctor`'s verdict is right, but no line names the offending key (finding 11) |
| M5 | Fixed | 3 allocations for 64 KiB of separators |
| m1 | Fixed | `Max-Age=2592000; HttpOnly; Secure; SameSite=Lax; Path=/` |
| m2 | Fixed on refresh, not at mint | See finding 12 (no lifetime-vs-cap check at mint) |
| m3 | Fixed | A 0755 leaf is refused |
| m4 | Fixed | Userinfo, query, fragment refused (echo in message is finding 10) |
| m5 | Fixed | Link expiry ceiling 1h, cap ceiling 2y, floors 1s |
| m6 | Fixed | `Store.Write` is gone |
| m7 | Fixed | `idRefusalReason` never prints the id |
| m8 | Fixed | One unreadable record no longer blocks startup |
| m9 | Fixed | Clear cookie renders `Max-Age=0` with matching attributes; the handler doc does not claim revocation |
| m10 | ASCII only | Non-ASCII control/format characters (NEL, CSI, RLO, NBSP, ZWSP) pass `NormalizeAddress` (finding 8) |
| m12 | Fixed | `RecordID` hashes a length-prefixed kind and value |
| m14 | Fixed | A panicking subscriber leaves `err=<nil>`; the next subscriber still runs |
| m11, m13 | Carried to 5d | As the plan states: the `ReadWritePaths` grant and the env template's nine keys |

**Disposition of this review's own nineteen new findings** (Geoff
ruled 2026-09-23): majors 1-4 plus minors 7, 8, 10, 11, 12, 13, 15, 17,
19 become 5c's security-fold Task 1; minor 18 (HSTS) joins 5c's served
header set; major 5 (trusted-proxy behind the tunnel) with minors 9 and
14 are 5d deploy gates, minor 14 landing with 5d's offboarding verb;
minor 16 (background-send load under the generic answer) is a recorded
residual, no fix required. M4's partial status is covered by finding
11 above; m10's ASCII-only status is covered by finding 8 above.

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
   design.
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
  `docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`.
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
  5b-i `HISTORY:` pointer in this plan sits 47 lines low at HEAD
  (Task 0 item 18, confirmed independently by this close):
  `:496-530` re-keys to `:543-577` (the 5b-i to 5b-ii handoff),
  `:533-535` to `:580-582` and `:540-543` to `:587-590` (both budgets),
  `:411-416` to `:458-463` (what this pass owes / the plan-prep tail).
  `:1692` (the 5a-flag citation at `:497`) is a rung 5a pointer, not a
  5b-i one, and is outside launch ruling (c)'s scope; it resolves
  correctly at its cited line and needs no correction.
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
write path** (`AddMember` refuses `Admin` on every channel but
`identity.ChannelServiceToken`), so the add-member handler inherits a
refusal rather than a rule to remember, and 5c's CLI grant supplies its
own caller with that one permitted channel; the verify-identity seam
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
sign-out; the security-fold Task 1 (security review majors 1-4 and the
listed minors); and the three rejoin-sentence corrections
(`CLAUDE.md:122`, `scripts/check.sh:5`, `Makefile:33`, the last worded
"The web legs rejoin", plural, present tense), `scripts/deploy.sh`'s
`config/contributors.yaml` delivery left alone by R37.

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
which the plan itself states do not count as a sitting).

---

*End of draft. Corrections above are drafted for the conductor's fold;
none has been applied to `docs/HISTORY.md`, `docs/STATUS.md`, or the
spec.*
