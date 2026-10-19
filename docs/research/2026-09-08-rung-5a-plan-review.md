# Adversarial review: the rung 5 pass 5a plan

Fresh-context review (2026-09-08) of
`plans/2026-09-08-rung-5a-serve-shell.md` revision 1, in two independent
lenses: ingress security, reading the plan as an attacker against the
shell it specifies, testing every spelling and encoding that could
survive the strip or the four normalization rules, what the fallthrough
returns, what the deadline policy leaves unbounded, and what the bind
surface exposes; and plan integrity, verified against the rung 5 spec
revision 2.1, the spec inventory, PA's serve-seam paragraph, the build
spec's B2, B3, B8 and B12, the registry contract in HISTORY's rung 1b
entry, the tree, and the 8a exemplar plan. Folded as revision 2 the same
day; the merged 30-item "Verified sound" list below is the
do-not-re-litigate record for the pass, alongside the rung 5 spec's own
records.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B3, M-a through M-j, m1 through m10.
Twenty-six raised findings became twenty-three after three merges. Every
finding was re-verified by the fold editor before folding, and every
mechanism claim about `net/http`, `net/url` or the kernel was measured on
the installed toolchain (go 1.27.1) rather than carried. Five reviewer
measurements did not survive that check in the form they were written and
are corrected below. One of the five retires a major to a minor; the
other four leave their findings standing. No finding was rejected
outright. Eight items are conductor-owed rather than folded silently, and
each is applied provisionally in the plan with an inline marker.

Both lenses agree the plan is well grounded where it reasons from the
spec and the inventory. The composition order matches PA:79 clause for
clause, every line of `spec:2035`'s 5a scope lands in a named task, all
six exit-gate assertions appear as acceptance criteria in the task that
lands the surface each grades, the registry contract is followed in the
words rung 1b wrote, the underscore-twin reasoning is exactly right about
`net/textproto`, R27's two-address split is correct including the
`net.SplitHostPort` requirement, the prior-art verdicts reproduce the
spec's table rather than inventing one, the independence marks are true
by files, and the plan stays inside 5a with a stop condition per
exclusion. The prose carries no em dash in 2,150 lines.

The defects concentrate in four places. The strip is specified over
`r.Header` alone, so an identity value arriving in a trailer survives it.
The one mechanism the spec does not specify is stated two ways in the
same resolution, one of which refuses nothing and the other of which
refuses every percent-encoded path. Three request shapes sit in the
pass-end reviewer's rubric and in no task's specification. And the
ceiling anchors on a pass that paid neither of the two costs this pass
mandates, on a row that is miscited.

---

## Corrections to the reviews' own evidence

**1. The security lens says `OPTIONS *` reaches the fallthrough with the
path `*`.** Measured, it reaches neither the fallthrough nor the shell's
own chain. `serverHandler.ServeHTTP` swaps in `globalOptionsHandler` for
`req.RequestURI == "*"` with method `OPTIONS` unless
`Server.DisableGeneralOptionsHandler` is set
(`$(go env GOROOT)/src/net/http/server.go:3409-3410`), and a live
`httptest` server answers such a request 200 with no handler invocation.
The finding stands on its other two shapes: `CONNECT` does reach the
handler with an empty `URL.Path` and a populated `URL.Host`, and an
absolute-form target routes on its path while setting `r.Host` to the
authority the client chose. The measurement adds a second consequence the
review did not claim, folded with it: `OPTIONS *` bypasses the ingress
strip along with everything else, so the plan's "the strip wraps
everything" is true of every request the chain sees and not of every
request the server accepts.

**2. The security lens grades the unbounded `Connection`-listed removal
as a major, on the ground that it lets a client delete the inputs to
5b's origin check.** The deletion is real and the exploit is not. The
Fetch standard lists `Connection` among the forbidden request-header
names, so no browser-driven cross-site request can carry an
attacker-chosen `Connection` value, and a client that can set arbitrary
headers on a victim's request has already defeated the origin check by
other means. Measured on the other half: Go's
`CrossOriginProtection.Check` returns nil when neither `Sec-Fetch-Site`
nor `Origin` is present (`net/http/csrf.go`, the `Check` method), so the
check does fail open on absent signals and the fold is still worth
taking. Retired to **m2** and folded as hygiene with a named keep-set,
rather than as an attack path.

**3. The security lens says the `IP_FREEBIND` mechanism is observable
only by reading the socket option back.** The boot property itself is
observable on this workstation, with no privilege. Measured: a
`net.ListenConfig` with no `Control` hook fails
`listen tcp4 192.0.2.77:0` with `bind: cannot assign requested address`,
and the same listen with `syscall.SetsockoptInt(fd, syscall.IPPROTO_IP,
syscall.IP_FREEBIND, 1)` in the hook succeeds and binds. `192.0.2.77` is
TEST-NET-1 and is on no interface here. The finding is upgraded rather
than corrected: Task 6's mutation 5 fires against a behavioral row, and
declared coverage gap 1 closes outright rather than narrowing. The
review's IPv6 concern is separately wrong: the same `IPPROTO_IP`
level and `IP_FREEBIND` constant succeed on an `AF_INET6` listener bound
to `[::1]`, reading back as 1, so no `IPV6_FREEBIND` constant is needed
and the loong64-only definition is irrelevant here.

**4. The integrity lens gives two tree pointers one line off in the
plan's favor.** `newRootCmd` is at `cmd/dubplate-server/main.go:63`, so
the plan's `:62` is off by one as the review says, but the `AddCommand`
block is `:74-79` exactly as the plan cites it and not `:75-80`. The
`isOptionalIntegrationKey` call site is `registry.go:273`, not `:272` as
the plan says and not `:272` as the review repeats. Both corrections
leave m7's drift list standing; only two of its rows move.

**5. The integrity lens miscites INV's budget table while correcting the
plan's citation of it.** Counted line by line, `INV:1111` is the 4b row,
`INV:1112` is 1b at 3.7M against 2.5M, and `INV:1113` is 8a at about
3.6M against 4.5M. The plan cites `INV:1111` for 8a and the review
proposes `INV:1112`. The correct pointer is `INV:1113`, and B3 folds that
one.

---

## BLOCKERS

### B1. The strip is specified over `r.Header` alone, so a trailer-borne identity header survives ingress and the pass's first claim is false as written

**Lens.** Ingress security.

**Plan lines.** Task 2's outcomes (`plan:1270-1299`) and acceptance
criteria (`plan:1320-1343`), which describe header removal and nothing
else; threat-model claim 1 (`plan:944-946`), which states that no inbound
identity header survives ingress on any path.

**Evidence.** The spec's strip section (`spec:765-816`) never mentions
trailers, and `grep -niE 'trailer'` over the spec returns nothing. The
plan's only occurrence is R-O's list of fixed RFC 7230 names
(`plan:906`). Measured against a live `httptest` server on go 1.27.1: a
chunked request carrying `Trailer: Remote-User` in its header section and
a `Remote-User: attacker` trailer after the body arrives at the handler
with `r.Header` empty of the name, `r.Trailer` carrying
`map[Remote-User:[]]` before the body is read, and
`map[Remote-User:[attacker]]` after it. Go's server moves announced
trailer keys into `r.Trailer` and removes the `Trailer` announcement from
`r.Header`, so neither group 1 nor group 3 can reach them: group 3 reads
`Connection`, and the announcement is not there to read.

**Consequence, built as written.** Every enumerated row in Task 2 and
every order assertion in Task 5 reads `r.Header`, so all of them pass
with the hole open. `httputil.ReverseProxy` forwards trailers to the
upstream (`reverseproxy.go:604-618`), so rung 12 hands the value to the
streamer. Exploitability at rung 12 needs a downstream that merges
trailer fields into its header section, which neither Go's server nor
`ReverseProxy` does today, so the immediate defect is a false claim
rather than a live bypass. The claim is what the pass-end
`web-auth-security-reviewer` grades against, and it will try this.

**Fix folded.** Task 2 gains a fourth clause: the strip clears `r.Trailer`
of every name in the strip set, and does so before the handler runs. Task
2 gains a row sending a chunked request that announces and carries
`Remote-User` as a trailer, asserting the terminal handler sees the name
in neither map, before and after reading the body. Task 5 gains the same
row against the composed chain. Task 2's mutation list gains "drop the
trailer clause", which fires on that row. Claim 1 in the threat model is
restated to name header and trailer fields, and the reviewer's attack
list gains the shape.

### B2. R-D states rule 1 two ways, one vacuous and one a blanket ban on percent-encoding, and no row in Task 3's table can tell them apart

**Lenses.** Both, independently. The integrity lens found the two
statements; the security lens found that the surviving reading refuses
every encoded path and that the plan never says so. One defect, folded
together.

**Plan lines.** R-D (`plan:711-727`), whose proposal at `:719` is "compare
`r.URL.EscapedPath()` against `r.URL.Path`" and whose rule statement at
`:722-724` is "refuses when a single decode of the escaped form does not
reproduce the decoded form byte for byte"; Task 3's first outcome
(`plan:1405-1407`); Task 3's required table (`plan:1428-1432`) and its
literal-status criterion (`plan:1434`).

**Evidence.** Measured with `url.ParseRequestURI` on go 1.27.1, thirteen
paths, one file, re-runnable. `PathUnescape(EscapedPath()) == Path` holds
for every one of them, `/a%2Fb`, `/a%5Cb`, `/a%2E%2E/b` and `/a%2520b`
included, because `URL.EscapedPath` returns `RawPath` only when
unescaping it reproduces `Path` and otherwise re-derives the escaped
form. The second reading therefore refuses nothing, and every smuggling
shape rule 1 exists to close survives it. The string comparison behaves
very differently: `EscapedPath() != Path` is true for `/a%20b`
(Path `/a b`), `/%E2%98%83` (Path `/☃`), `/a%41b` (Path `/aAb`), `/%00`
and `/%0A`, so it is a blanket ban on percent-encoding in a path rather
than a rule against encoded separators. The plan never states which it
means. Task 3's table carries refusal rows plus one plain passing path,
and both readings pass it.

**Consequence, built as written.** The implementer picks the shell's
path-refusal policy, and the choice is adopter-visible on every path rung
12 proxies. A library whose track paths carry a percent-encoded UTF-8
byte is unreachable through the shell under one reading. Under the other,
`%2F` reaches the prefix match. The choice also decides whether `%00` and
`%0A` are refused, since rule 1 is the only rule that reaches them.
R-D's stated reason for declining `RawPath` is sound but incomplete: it
argues against a `RawPath == Path` comparison and never considers
`RawPath != ""`, which is the idiomatic non-default-encoding test.
Measured, that test is narrower than either reading: `/a%2Fb`,
`/a%2E%2E/b` and `/a%41b` populate `RawPath`, while `/a%5Cb` and `/a%20b`
do not, because the default encoder escapes a backslash and a space
itself.

**Fix folded, conductor-owed.** R-D states rule 1 once, as a single
predicate, and the "single decode reproduces the decoded form byte for
byte" sentence is deleted from R-D and from Task 3's outcomes. The policy
is stated in C3 in one sentence, applied provisionally as the blanket
form: a path carries no percent-encoding at all, on the ground that the
shell owns no path that needs one and rung 12's driver declares its
bypass paths as literals. Task 3's table gains a benign encoded byte
(`/%E2%98%83`) with its verdict written out, so the row fails under the
other policy. The forward criterion (`plan:214-223`) records the
consequence for rung 12, and the close's spec corrections carry the
policy to `spec:918-920`, which states the rule and not the mechanism.

### B3. The ceiling anchors on a miscited row, and omits the two costs this pass mandates that its anchor pass never paid

**Lens.** Plan integrity.

**Plan lines.** The ceiling table (`plan:246-252`) and the headroom
paragraph (`plan:254-266`).

**Evidence.** The anchor row cites `INV:1111` for 8a's 3.60M. Counted
line by line, `INV:1111` is 4b, `INV:1112` is 1b, and `INV:1113` is 8a
(correction 5 above). Two costs 5a mandates are absent from 8a's actual.
`HISTORY:367-368` records 8a's forward carry 1, "`code-simplifier` did
not run over 8a's Go", and 8a's own plan states "This pass runs in the
MAIN TREE", so its 3.6M carries no per-task simplifier and no merge
ritual. 8a's spend line breaks down as prep 0.87M, Task 0 0.1M, chain
2.27M across 24 agents, leftovers 0.2M, conductor turns 0.15M
(`HISTORY:405-409`), with no simplify entry. This plan mandates both in
conductor actions 1 and 3. `HISTORY:1719` prices 4a-i's simplify at 0.15M
with its review at about 0.05M.

**Consequence, built as written.** Adding 0.20M for the simplifier and
its review, plus the merge ritual's two extra full-gate runs and its
conditional post-rebase `diff-reviewer`, moves the expectation from 3.71M
to about 3.96M against a proposed 4M ceiling. The stated headroom of
0.29M is nearer 0.04M, on a pass the plan itself groups with 1b, which
ran 3.7M against 2.5M and crossed its flag inside the chain with no
conductor action. This is the defect 8a's own review filed as its B3, in
the same shape.

**Fix folded.** The anchor citation becomes `INV:1113`. The table gains
two lines, the per-task `code-simplifier` with its review at +0.20M and
the merge ritual at +0.05M, and one for the review folds themselves,
which add rows to six tasks, at +0.10M. The expected total becomes 4.06M
and the ceiling is proposed at **4.25M**, with the 80% flag at 3.4M. The
rung total against R23's 14.5M reservation becomes 15.75M, recorded
rather than taken from another pass. The flag paragraph keeps its named
action and gains a second: at 3.4M the conductor chooses between running
the pass-end reviewer here and handing the read to 5b.

---

## MAJORS

### M-a. R-O's adopted rule strips `Upgrade` from the exact request R-O declines to break

**Lens.** Plan integrity.

**Plan lines.** R-O (`plan:902-919`), which declines removing the fixed
RFC 7230 set because "a shell that also removed them would strip
`Upgrade` from a websocket request the streamer's own player may need",
then adopts "5a removes what `Connection` lists and nothing else"; Task
2's third group outcome (`plan:1290-1292`) and its `Connection` criteria
(`plan:1333-1335`).

**Evidence.** A websocket handshake sends `Connection: Upgrade` alongside
`Upgrade: websocket`, so the adopted rule deletes `Upgrade` on precisely
the request R-O protects. Go's own reverse proxy shows both the hazard
and the carve-out: it reads `upgradeType(outreq.Header)` at
`httputil/reverseproxy.go:464`, calls `removeHopByHopHeaders` at `:469`,
and re-sets `Connection: Upgrade` and `Upgrade: <type>` at `:481-484`
under the comment "add back any necessary for protocol upgrades, such as
for websockets". `removeHopByHopHeaders` itself does both halves
(`:646-661`), the `Connection`-listed loop first and the fixed
`hopHeaders` list second. None of Task 2's six mutations reaches an
upgrade request, so nothing in the pass fires on it. The plan also never
says whether the `Connection` header itself survives the strip.

**Fix folded.** R-O decides the upgrade case explicitly rather than
reasoning from the fixed-set half alone, citing `ReverseProxy`'s re-add
as the precedent, and states what happens to the `Connection` header
itself. Applied provisionally: a request whose `Connection` names
`Upgrade` keeps `Upgrade` and keeps `Connection`, and every other listed
name is removed. Task 2 gains a `Connection: Upgrade, Remote-User` row
asserting `Remote-User` gone and `Upgrade` intact, and a mutation
"remove every listed name unconditionally", which fails that row.

### M-b. Group 2 covers three forwarded headers and omits the `X-Forwarded-Host`, `-Proto` and `-Prefix` family

**Lens.** Ingress security.

**Plan lines.** C2's group 2 (`plan:496-500`) and Task 2's group 2
outcome (`plan:1287-1289`), both naming exactly `X-Forwarded-For`,
`X-Real-IP` and `Forwarded`.

**Evidence.** `spec:795-800` names the same three and no more. B12 asks
for proxy hygiene as a class, and the Authelia forwarded-headers
documentation Task 2 already consults covers the whole family.
`X-Forwarded-Host`, `X-Forwarded-Proto` and `X-Forwarded-Prefix` pass the
shell untouched today. They reach the streamer through rung 12's
fallthrough, where they steer any absolute URL a downstream builds, and
they reach 5b's handlers, where any base-URL or scheme derivation would
read a client-asserted value. B2's absolute-form measurement (M-c) makes
the same point about `r.Host` through a different door.

**Fix folded.** Group 2 extends to the `X-Forwarded-*` family under the
same trusted-proxy carve-out, since a trusted front legitimately sets
them and the derivation's carve-out is already keyed on the peer address.
Task 2's citation work covers each added member, and Task 2's enumerated
expected set carries the family so a missing name fails. The forward
criterion records the wider set as what rung 12 inherits.

### M-c. Three non-origin-form request shapes sit in the reviewer's rubric and in no task's specification

**Lenses.** Both. The security lens found the rules pass vacuously; the
integrity lens found the table row with no verdict. One defect.

**Plan lines.** The threat model's request-target attacks
(`plan:977-983`); Task 3's table, which lists "an absolute-form request
target" (`plan:1432`) with no expected verdict against a criterion
requiring every expected status to be a literal in the table
(`plan:1434`).

**Evidence.** Measured against a live server on go 1.27.1. A `CONNECT`
request arrives with `URL.Path` empty, `URL.Host` set to the authority
and `RequestURI` in authority form, so all five normalization rules pass
vacuously and it reaches the fallthrough. An absolute-form `GET
http://evil.example/dubplate/x` arrives with `URL.Path` `/dubplate/x`,
`URL.Host` `evil.example` and `r.Host` `evil.example`, so it routes on
its path like an origin-form request while handing every downstream an
authority the client chose. `OPTIONS *` never reaches the chain at all
(correction 1). The plan's criterion at `plan:1434` cannot be satisfied
against a row whose verdict no rule in the plan decides, so the
implementer would write the behavior and its expectation together, which
is the derived-expectation shape the Global Constraints forbid.

**Fix folded, conductor-owed.** C3 gains a sixth rule: a request whose
`URL.Path` does not begin with a separator is refused, answered with the
same status as rules 1 through 4. Task 3's table gains verdicts for
`CONNECT` and for an absolute-form target, and Task 5's composed-chain
assertions gain both. Applied provisionally, an absolute-form target
routes on its path like any other, because refusing it would break an
adopter's own proxy, and the residual is recorded instead: `r.Host` is
client-asserted on that shape, which 5b must not read as an origin. Task
3 gains the mutation "drop the origin-form rule", which fires on the
`CONNECT` row. The `OPTIONS *` bypass is recorded as a residual in the
threat model rather than as a rule, since the stdlib answers it and
`DisableGeneralOptionsHandler` is a decision 5a has no reason to take.

### M-d. Rule 3 refuses any empty segment, which contradicts Task 4's own `/dubplate/` row and leaves the root path undecided

**Lens.** Ingress security.

**Plan lines.** C3 rule 3 (`plan:530`), Task 3's outcome (`plan:1409`)
and its table row "a trailing empty segment" (`plan:1430`) with no
verdict; Task 4's table row "the prefix with a separator"
(`plan:1529-1530`), which is expected to enter the owned surface.

**Evidence.** Under a naive segment split, `/dubplate/` yields a trailing
empty segment and `/` yields a leading and a trailing one, so a literal
reading of "any empty segment" refuses both. Task 4's table then
contradicts Task 3's. Rule 3's own gloss says "a doubled separator",
which is the narrower intent, and the rule's name is what an implementer
builds to. Measured on the mux for the adjacent behavior: `/dubplate//x`
draws a 307 to `/dubplate/x` from `http.ServeMux`, which is the cleaning
rule 3 exists to preempt, while `/dubplate/` draws a plain 404 with no
redirect.

**Fix folded.** Rule 3 is restated as "a doubled separator anywhere in
the path", with two sentences saying that the leading separator is not a
segment and that a single trailing separator is admitted. The verdicts
for `/`, `/dubplate/` and `//dubplate` are written into both Task 3's and
Task 4's tables. The root path is servable and reaches the fallthrough,
which follows from B3's one-owned-prefix rule and needs stating because
rule 3 is the only thing that could have refused it.

### M-e. The deadline policy sets no read timeout, so a slow request body holds a connection, and the plan mis-describes the salvage server it adopts from

**Lens.** Ingress security.

**Plan lines.** C6 (`plan:596-601`), Task 6's timeout outcome
(`plan:1716-1718`), declared coverage gap 6 (`plan:1968-1971`).

**Evidence.** `ReadHeaderTimeout` stops covering the connection once the
headers are read, `IdleTimeout` applies between requests, and
`WriteTimeout` bounds writes. Nothing in the named set bounds a body
trickled a byte at a time, and 5b's login and consumption POSTs take
bodies. The salvage server this pass adopts from sets four timeouts, not
one: `httpReadHeaderTimeout = 10s`, `httpReadTimeout = 30s`,
`httpWriteTimeout = 30s` and `httpIdleTimeout = 120s` at
`_salvage/internal/serve/serve.go:599-602`, all four applied at
`:630-633`. C6's opening sentence says the salvage server "sets one
global 30-second write timeout", which drops the read timeout it also
carries, so the plan rejects a policy its adoption material does not
have. Gap 6 then rests slowloris resistance on the timeout constants,
which covers the header half alone.

**Fix folded, conductor-owed on the value.** C6's salvage sentence is
corrected to name all four constants and to say that the write timeout
alone is what B12 answers. Task 6's constant set gains a read timeout
with its own reason comment, applied provisionally at the salvage
server's 30 seconds, and the acceptance criteria name it. The 5b handoff
carries the alternative in case the conductor declines the constant: bound
the body with `http.MaxBytesReader` and extend with
`ResponseController.SetReadDeadline` per route. Gap 6 is corrected to
name the half the constants answer.

### M-f. Task 5's fourth order assertion cannot be implemented against the production chain

**Lens.** Ingress security.

**Plan lines.** Task 5's fourth order assertion (`plan:1638-1642`), the
Global Constraint forbidding structural proof (`plan:344-347`), Task 5's
mutation 1 (`plan:1659-1662`).

**Evidence.** The assertion requires proving strip-before-normalization
by sending a request carrying both an identity header and a `..` segment
and showing the header was already gone when the refusal happened,
"proved by a normalization step that records what it saw". A request
refused by normalization reaches no handler, so nothing outside the chain
observes what normalization saw. The two available implementations are to
rebuild the chain inside the test, which asserts on construction and is
what the Global Constraint forbids, or to add a test-only seam the
production path does not use. This is the pass's headline assertion for
risk 1 and the only one that catches Task 5's mutation 1.

**Fix folded.** The refusal path gains a structured log record naming
whether any strip-set spelling was present on the refused request, and
the assertion reads it through a test logger against the production root
handler. The mechanism is stated in the acceptance criterion rather than
left to the implementer. The record is also what R-N's "the distinction
lives in the log record" already promised for the terminator and the
empty fallthrough, so the two share one mechanism.

### M-g. The right-to-left parse property is mis-stated, and its own acceptance fixture contradicts it

**Lens.** Ingress security.

**Plan lines.** C7's third bullet (`plan:621-623`) and Task 1's matching
outcome (`plan:1183-1186`), both ending "so an attacker prepending
entries cannot choose the result"; Task 1's criterion (`plan:1202-1205`).

**Evidence.** The criterion fixes a three-hop value whose two rightmost
entries are configured proxies and whose leftmost is attacker-supplied.
Taking the rightmost hop that is not a configured proxy returns that
leftmost value, so in the plan's own fixture the attacker did choose the
result. The property that actually holds is narrower: the answer is the
rightmost hop that is not a configured proxy, so an attacker cannot reach
past the trusted chain by prepending entries, and the guarantee rests on
every configured proxy appending the address it observed rather than
forwarding what it was handed. The review places this property in the
pass-end reviewer's rubric; it is not there. C7 is the implementer's and
the `diff-reviewer`'s contract, which is where the mis-statement does its
damage.

**Fix folded.** C7 and Task 1 state the property as "the rightmost hop
that is not a configured proxy, so an attacker cannot reach past the
trusted chain", and both record the residual that the derived value is
attacker-controlled whenever the attacker is the nearest non-proxy hop.
The fixture is correct and stays; the restated property is what justifies
its expected value.

### M-h. Task 6 pre-authorizes an unfired mutation for `IP_FREEBIND`, and both the option and the boot property are directly observable

**Lens.** Ingress security, with the fold editor's measurement upgrading
it (correction 3).

**Plan lines.** Task 6's `Control`-hook criterion (`plan:1757-1760`),
which permits the assertion to "report itself unfired with the reason";
mutation 5 (`plan:1777-1781`); declared coverage gap 1
(`plan:1952-1955`).

**Evidence.** Measured on this workstation, unprivileged. Without a
`Control` hook, `lc.Listen(ctx, "tcp4", "192.0.2.77:0")` fails
`bind: cannot assign requested address`. With
`syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_FREEBIND, 1)`
in the hook, the same listen succeeds and binds. `192.0.2.77` is
TEST-NET-1 and is on no interface here, confirmed against `ip -4 addr`.
`syscall.GetsockoptInt` reads the option back as 1 in the same hook.
`syscall.IP_FREEBIND = 0xf` is present at
`$(go env GOROOT)/src/syscall/zerrors_linux_amd64.go:433`. The same level
and constant succeed on an `AF_INET6` listener bound to `[::1]`.

**Fix folded.** Task 6 asserts the property behaviorally: a listener set
containing a documentation-reserved address that is on no interface opens
successfully, and a sibling row without the hook fails to bind. Mutation
5 becomes a firing mutation. Declared coverage gap 1 is deleted rather
than reworded, which moves the pass's gap count from seven to six and the
running total from forty-three to forty-nine. Task 6 gains an IPv6 bind
row, and its outcomes state that the hook sets `IP_FREEBIND` at
`IPPROTO_IP` for both address families.

### M-i. Task 2's criteria and two of its mutations need artifacts Tasks 4 and 5 create

**Lens.** Plan integrity.

**Plan lines.** Task 2's Interfaces block, which produces "a wrapper
taking an `http.Handler` and returning one"; its criteria at
`plan:1328-1329` ("on a path under the prefix and on the fallthrough")
and `plan:1338-1340` ("The exemption's test constructs a shell with a
front configured"); mutation 2 (`plan:1349`), which resolves to "the
fallthrough row fails".

**Evidence.** The owned prefix and its terminator are Task 4's, in
`internal/serve/prefix.go`. The shell's construction value and the
fallthrough slot are Task 5's, in `internal/serve/serve.go`. At Task 2
the implementer either stalls or invents prefix and fallthrough terminals
inside `strip_test.go`, which proves nothing about the composed chain.
The genuine assertion already sits in Task 5's criteria, where
`spec:938-941`'s three order assertions live.

**Fix folded.** Task 2's criteria are restated against the strip wrapper
and two recording test handlers, the exemption row names the strip's own
construction value rather than "a shell", and the "under the prefix and
on the fallthrough" assertion moves to Task 5 with mutation 2's
fallthrough row. Task 2 keeps the enumeration, the underscore proof, the
citation work and the group 2 and group 3 rows, which are all reachable
from the wrapper alone.

### M-j. The pass's headline trust claim is held by a read-the-code criterion in the task that lands it

**Lens.** Plan integrity.

**Plan lines.** Task 1's criterion (`plan:1197-1198`), "The peer-address
function reads `r.RemoteAddr` and no header, provable by reading the
function in full"; Task 1's mutation 4 (`plan:1265-1268`), admitted to
fire only in Tasks 2 and 5.

**Evidence.** Threat-model claim 3 is "No configured header can change a
trust decision, because every trust decision reads the peer address", and
the pass-end reviewer grades against that section. Task 1's only
criterion for it is structural. The plan's Global Constraint that every
ordering property is proved by behavior does not literally reach this
claim, which is not an ordering property, and the substance is the same:
the task that lands the shell's central guarantee has no row that fails
when the guarantee is broken, and its own mutation table says so.

**Fix folded.** Task 1 gains one table row: mode on, an allowlisted peer,
and a client-address header naming a different address, asserting the
peer-address function returns `RemoteAddr`'s host. Mutation 4 then fires
in its own task rather than as a cross-task promise, and its stated blast
radius stays as written.

---

## MINORS

**m1. The refusal status for normalization rules 1 through 4 is never
named.** C3 says each rule is "a refusal with a status" (`plan:520`) and
Task 3 says "the refusal status is stated once and used by every rule"
(`plan:1420`), and no status literal for those rules appears anywhere.
`grep -nE '\b(400|404|414|501)\b'` over the plan returns 414 for the
length bound and 404 for the terminator and the empty fallthrough only,
and no 400 at all. Task 3's criterion requiring every expected status to
be a literal in the table cannot be met from the plan, so the implementer
picks a security-visible value. Folded by naming `400 Bad Request` in C3,
provisionally, and by keeping the 414 exception with the sentence it was
missing: a length refusal is distinguishable because the shape is
distinguishable from outside anyway, and the indistinguishability goal
covers the four smuggling rules.

**m2. Group 3's `Connection`-listed removal has no keep-set.** The rule
removes every header the request's own `Connection` header lists, with no
exclusions (`plan:1290-1292`, C2 at `plan:502-508`, `spec:802-810`), so a
client can delete `Cookie`, `Origin` or `Sec-Fetch-Site` from its own
request before anything downstream reads them. Measured, Go's
`CrossOriginProtection.Check` returns nil when neither `Sec-Fetch-Site`
nor `Origin` is present, so 5b's origin check does fail open on absent
signals. The attack path is closed by the Fetch standard, which lists
`Connection` among the forbidden request-header names, so no browser can
set it cross-site (correction 2). Folded anyway, because it costs one
line: group 3 carries a named keep-set of `Cookie`, `Origin`,
`Sec-Fetch-Site` and the shell's own session-cookie name, with a Task 2
row asserting `Connection: Origin` leaves `Origin` standing, and the R-F
handoff records that 5b's check sees stdlib fail-open behavior on absent
signals whatever the shell does.

**m3. The strip exemption says nothing about a duplicate value of the
exempted header.** C2's exemption (`plan:487-494`) and Task 2's exemption
row (`plan:1338-1340`) cover one named header surviving for a matching
peer and its underscore twin not surviving. A request arriving with two
values of the exempted name is what a front that appends rather than
replaces produces, and `Header.Get` returns the first. The exemption is
the one path on which any identity header survives the shell, and PA:80
closes `internal/serve` to other lanes after rung 5, so the shape is
fixed here. Folded as: the exemption collapses the exempted header to its
last value, with the test row, and the forward criterion records it so
rung 12 and any future front inherit the rule.

**m4. A single dot segment is admitted while `..` is refused.** C3's four
rules reach no `.` segment (`plan:518-536`). Measured, `http.ServeMux`
answers `/dubplate/./x` with a 307 to `/dubplate/x`, which is exactly the
cleaning behavior rule 2 exists to refuse, and the fallthrough forwards
`/rest/./x` to a streamer whose router may normalize differently. Folded
by refusing a `.` segment alongside `..`, as part of rule 2 rather than
as a sixth rule, with the row.

**m5. The session-cookie removal has no criterion covering the rest of
the `Cookie` header.** Raised by both lenses. Task 5's criterion asserts
the cookie is absent at the fallthrough and present under the owned
prefix (`plan:1646-1648`). `r.Header.Del("Cookie")` at the fallthrough
satisfies both assertions while removing the streamer's own cookies from
every proxied request, which surfaces at rung 12 as a broken streamer
session with a green 5a suite behind it. No criterion says the other
cookies survive byte for byte, or that every duplicate of the named
cookie is removed rather than the first. Folded as two rows, a request
carrying three cookies including the named one and a request carrying the
named cookie twice, plus the mutation "delete the whole `Cookie` header",
which fails the first.

**m6. R-G's live-mode descriptor is still derived from configuration.**
R-G argues that "a doctor that recomputed the mode from configuration
would report what was configured, not what runs" (`plan:770-776`), and
Task 1 produces the descriptor from the same constructor arguments the
mode is configured with (`plan:1161-1166`). The descriptor proves that
configuration reached the shell, which is one hop better than re-reading
the environment and is not evidence that any request was classified.
Folded by saying exactly that in R-G and in declared gap 4, so 5d reads
the descriptor as the shell's own constructed state and puts the
request-path evidence in its probe row.

**m7. Fourteen spec, INV and tree pointers drift, including four the plan
reasons from.** Re-resolved by the fold editor against the tree, with the
review's own two errors corrected (correction 4). Hook 6 is `spec:1843`,
cited as `spec:1845` in four places. `internal/serve`'s package contract
is `spec:1474-1479`, cited as `spec:1479-1486` twice. Risk 1's
behavior-not-structure clause is `spec:2224-2226`, cited as
`spec:2227-2229` twice. Risk 11 is `spec:2281-2286`, cited as
`spec:2277-2283`. The "same one owned prefix" sentence is `spec:494-495`,
cited as `spec:485-489`. INV's 8a budget row is `INV:1113`, cited as
`INV:1111` (B3). INV's prep pattern is `INV:1117-1121`, cited as
`INV:1117-1123`, and its overrun pattern `INV:1123-1128`, cited as
`INV:1125-1131`. `newRootCmd` is `main.go:63`, cited as `:62`. The
Registry collision comment is `registry.go:24-31`, cited as `:23-30` and
`:27-30`. `isOptionalIntegrationKey` is `registry.go:284`, called from
`:273`, cited as `:279-291` and `:272`. `requireSameOrigin` is
`_salvage/internal/serve/serve.go:152`, cited as "around `:130`". The
`unix.SetsockoptInt` call is `:570`, cited as `:573`. The `keys.go` serve
class comment is `:75-78`, cited as `:74-79`. Content matches in every
case; only the ranges moved. Folded by correcting all fourteen, and by
adding one sentence to Task 0: a pointer off by fewer than about ten
lines whose cited content matches is a note rather than a blocker, so the
blocker channel stays for real staleness.

**m8. 117 of 204 top-level bullets open with a bolded lead phrase.**
Measured with `grep -c '^[-*] \*\*'` against `grep -c '^[-*] '`, 117 of
204 here (57%) against 38 of 212 (18%) in the 8a exemplar. Total `**`
runs are 1,156 against the exemplar's 768 in a slightly longer file. The
writing-voice shape rule reads "Do not open every bullet or paragraph
with a bolded lead phrase. Sparingly it signposts; by reflex it is the
machine list default." Em dashes are clean at zero and no scaffold header
appears. Folded by unbolding the routine list items in the per-task
Outcomes, acceptance-criteria and mutation lists, keeping bold where the
lead phrase is the item's name or a stop condition. Measured after the
fold: 50 of 221 top-level bullets (23%), with the survivors concentrated
in the Global Constraints, the threat model's attack list and the
contract sections, where the lead phrase names the item. `tellgrader`
reports 0.58 tells per 1,000 words over 20,842 words at a cadence CV of
0.77, and every flagged token is the literal word "underscore", which the
slop lexicon reads as a verb.

**m9. Fifteen resolutions stay provisional through execution, and no task
stamps the markers the close reads.** `plan:651` makes each resolution
"provisional, conductor until ruled", and `plan:1944-1947` requires the
close to list all fifteen "with their markers", each row naming the file
and the symbol the marker sits on. No task is told to write one. The
convention exists in the tree as `// ruled, conductor 2026-09-07, R1` at
`internal/config/registry.go:33`, `keys.go:35`, `main.go:62`,
`shadow.go:27` and `doctor_test.go:150`. The 8a exemplar closed its
resolutions before execution instead: "All ten stand as applied, ruled by
the conductor 2026-09-08". R-A and R-B compile into a registry
declaration whose `validateThresholdDefault` refuses an empty
`KindString` default (`registry.go:320-322`), so a ruling that changes
either after Task 7 re-edits a landed contract file and its test
literals. Folded by ruling the fifteen at dispatch as 8a did, and by one
line in the Global Constraints telling each task to stamp
`// ruled, conductor 2026-09-08, R-x` on the symbol its resolution
decides.

**m10. R33 is named as a ruling 5a reads and is never disposed of.**
`plan:84-85` lists R6, R15, R20, R22, R26, R27, R32 and R33 as the ones
5a reads. R32 is disposed once, at `plan:115`, as 5c's. R33
(`spec:674-686`) states a single-decode path predicate that reads
adjacent to Task 3's rules, and it appears nowhere in the plan body, so a
reader cannot tell whether it constrains 5a's normalization. Folded by
one sentence in R-F: R33 governs the post-login destination at 5b and
imposes nothing on 5a's path rules, which is what `spec:679-686` says
when read whole.

---

## Rejected findings

None retired. Every finding in both reviews survived re-verification.
One major was re-graded to a minor on a measurement that closed its
attack path (m2, correction 2), one major was upgraded on a measurement
that made its fix cheaper and stronger (M-h, correction 3), and three
reviewer measurements were corrected without moving their findings
(corrections 1, 4 and 5).

---

## Verified sound

Do not re-litigate these. Each was checked by the fold editor against the
named artifact, and every mechanism claim was measured on go 1.27.1.

1. The composition order matches PA:79 clause for clause: ingress strip,
   path normalization, prefix routing, fallthrough, with normalization
   wrapping the mux rather than registering as a route (C1 against PA:79
   and `spec:907-936`).
2. Risk 1's premise is real and the answer is right. Measured,
   `http.ServeMux` answers `/dubplate/./x`, `/dubplate//x` and
   `/dubplate/y/../x` with a 307 to `/dubplate/x`, so a check registered
   as a route runs after the decision it guards. Wrapping the mux is the
   correct answer, and Task 3's mutation 5 is the right way to prove it.
3. The underscore-twin reasoning. `textproto.CanonicalMIMEHeaderKey`
   uppercases the letter after each dash and treats the underscore as an
   ordinary token character, so `Remote-User` and `Remote_user` are
   distinct map keys, and `Header.Del` canonicalizes its argument, so a
   twin generated by replacing dashes with underscores deletes correctly
   without the plan spelling the canonical form. Task 2's requirement to
   assert it against the stdlib rather than remember it is right.
4. Segment-aware, case-sensitive prefix matching, and the salvage
   evidence behind it. `_salvage/internal/serve/serve.go:95` documents
   that `http.StripPrefix` does a plain string trim, so an unguarded
   `/uploadfoo` trims and passes. Refusing rather than documenting that
   class is sound, and `/Dubplate` reaching the fallthrough is the right
   reading of `spec:2035`'s exit gate.
5. The terminator. An unknown path under the owned prefix answering 404
   without reaching the fallthrough (C4, `spec:894-901`) is what keeps
   rung 14's API from being captured by rung 12's proxy, and Task 4's
   three-recording-handler shape proves it rather than inferring it from
   a status.
6. R27's two-address split. The peer address is `RemoteAddr` with the
   port stripped, it is the only input to any trust decision, the
   configured client-address header is read only when the peer matches an
   allowlisted proxy, and a malformed or empty value falls back to the
   peer rather than to an empty limiter key. Requiring
   `net.SplitHostPort` and an IPv6 row is correct.
7. R-M's dependency call. `syscall.IP_FREEBIND = 0xf` is present at
   `zerrors_linux_amd64.go:433`, and both the set and the readback work
   through `syscall` alone, so the IPv4 and IPv6 legs need no
   `golang.org/x/sys` and R26's budget of two dependencies, both 5b's,
   holds.
8. The extension point and the driver's point are stated as PA:79
   requires: a registration seam under the owned prefix plus exactly one
   root-level fallthrough slot, with rung 12's proxy occupying the
   fallthrough and `Mount(prefix, http.Handler)` declined for PA:77's
   reason (C5, R-L, Task 5's doc-comment criterion).
9. Every clause of `spec:2035`'s 5a scope lands in a named task: R27's
   two addresses (Task 1), the three-group strip with hop-by-hop removal
   (Task 2), the normalization rules (Task 3), the segment-aware
   case-sensitive match and the terminator (Task 4), the extension point
   and the empty fallthrough (Task 5), the deadline policy, the
   `ResponseController` mechanism and `IP_FREEBIND` (Task 6), the three
   keys with the reserved class retired and `newServeCmd` wired (Task 7).
10. All six of `spec:2035`'s exit-gate assertions appear as task
    acceptance criteria, each in the task that lands the surface it
    grades.
11. The registry contract is followed as `HISTORY:493-498` words it: the
    three keys are declared in `cmd/dubplate-server/serve.go` on rung 4's
    `shadow.go` precedent, and the `serve` reserved class is deleted from
    `internal/config/keys.go` in the same commit, with the criterion that
    the `keys.go` diff is a deletion and nothing else.
12. Task 7's mutation 1 is sound. `Registry.claim` is shared across
    families and `panicDuplicate` fires (`registry.go:95`, `:119-121`,
    `:270`), so declaring the thresholds without deleting the class
    panics at registration and the gate-resident registry test reports
    it.
13. Task 0 item 8's doctor arithmetic is right. `newDoctorCmd` prints one
    line per reserved class regardless of key count
    (`doctor.go:40-45`), so removing the serve class and adding three
    thresholds is net plus two, and both landed assertions derive from
    the registry rather than from a literal
    (`doctor_test.go:129-131`, `TestDoctorCmd_LineCountFollowsRegistry`
    at `:151-163`), so neither breaks.
14. R-D's stated trap is real. Measured, `RawPath` is empty for
    `/dubplate/x`, `/a%5Cb` and `/a%20b`, so a `RawPath`-against-`Path`
    comparison would refuse every ordinary path. Only the rule's
    statement is defective (B2).
15. The `keys.go` serve class is at `:74-85` with the "a bind address is
    not an identity concern" comment the plan carries forward, and the
    `auth and identity` class above it is untouched by 5a, as Task 0 item
    4 expects.
16. `main.go`'s package doc comment already lists
    "serve: run the HTTP server (rung 5)" at `:12`, so Task 7 adds a verb
    the doc comment names and needs no amendment there.
17. R-K. `spec:2039` and `spec:2042` give the serve unit, its
    `ReadWritePaths` grant and the `deploy.sh` pre-creates to 5d, so
    writing no systemd content and carrying no `vps-conventions` material
    is correct for this pass.
18. R-E's placement. The session-cookie removal belongs to the shell
    (`spec:808-813`), no pass row in the spec's split table claims it,
    and putting it at the fallthrough boundary with the cookie name
    supplied at construction is the reading that leaves 5b an edit-free
    inheritance.
19. R-N. 404 rather than 501 for an empty fallthrough, on the ground that
    501 advertises a capability the deployment lacks, and the terminator
    being indistinguishable from the empty fallthrough from outside as a
    property rather than a defect.
20. R-O's fixed-set half. Declining to remove the RFC 7230 fixed set at
    ingress is right: `httputil.ReverseProxy` removes them on the way out
    at `:469`. Only the upgrade case inside the adopted half is defective
    (M-a).
21. The deadline extension shape. `http.ResponseController.SetWriteDeadline`
    is the mechanism, and Task 6's paired rows, a handler that extends and
    succeeds beside a sibling that does not and fails, prove the default
    is real as well as extensible, which is what `spec:978-981` asks for.
22. The listener set's failure handling. Closing every listener already
    opened before returning on a partial bind failure matches
    `_salvage/internal/serve/serve.go:582-596` and is the right property
    to assert.
23. Task 0's shape. Resolving every spec, INV, PA, BS and PRA pointer
    against the real HEAD, treating a moved citation as a finding, and
    using the lane registry rather than `pgrep -f Projects/dubplate` as
    the live-executor check are all correct, and item 8's line-count
    arithmetic is properly made a blocker rather than an assumption.
24. The worktree and merge-ritual rules are stated in three places, the
    lane registry is named as the live-executor check, and the four
    surfaces that could make the rebase non-fast-forward are named and
    checked by Task 0 item 3. The `AddCommand` contention is named as a
    one-line edit resolved by the merge ritual, in the contested-surfaces
    table, the territory wall, Task 0 item 7 and Task 7's Files list.
25. The plan stays inside 5a. No token, cookie construction, roster,
    limiter, origin check, page, mail, systemd unit or deploy path is
    built, each exclusion carries a stop-and-report instruction, and the
    session-cookie removal it does take is the spec's own shell
    assignment drafted as a spec correction because no pass row claims
    it.
26. No module dependency is added. R-M routes `IP_FREEBIND` through
    `syscall` inside a `net.ListenConfig.Control` hook, and Task 6
    carries a `go.mod`-unchanged criterion plus a grep for
    `golang.org/x/sys`, keeping R26's two-dependency budget wholly with
    5b.
27. The opus upshift decision is made and justified: normalization takes
    it because its rules are specified and its mechanism is not, the
    strip stays `sonnet` because its mechanism is a map delete over
    generated data with a stdlib behavior the spec states exactly, and
    Task 5 carries a conditional upshift tied to Task 3's review verdict.
    B2 is the evidence the upshift was called correctly.
28. The 2.0 hook accounting is right against `spec:1843`: hook 6 is
    neutral and kept open, with the strip-as-data half landed at Task 2
    and the empty-fallthrough half at Task 5; hook 5 is 5b's and 5c's;
    every other task states neutral with a reason. Only the pointer was
    off (m7).
29. Deliverable counts are stated per task and none passes four; the
    checkpoints land on the plan's own four-task interval; the
    independence marks are true by files, with Tasks 1, 3 and 4 writing
    `clientaddr.go`, `normalize.go` and `prefix.go` and consuming nothing
    from each other; and the plan is right that they still run
    sequentially because one tree runs one executor.
30. The pass-end review shape matches CLAUDE.md and `spec:2326-2351`:
    in-chain `diff-reviewer` per task, one fresh-context
    `web-auth-security-reviewer` at the pass end with the threat model as
    its rubric and the task decomposition withheld, and `code-simplifier`
    dispatched by the conductor rather than by a subagent. No calendar
    schedule and no launch date appears anywhere in the plan.
