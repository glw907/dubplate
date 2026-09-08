# Rung 5a security review: findings and dispositions

Fresh-context `web-auth-security-reviewer` (`claude-opus-5`) read the
whole serve shell against `docs/superpowers/plans/2026-09-08-rung-5a-serve-shell.md`'s
threat-model section and the rung 5 spec's own threat model
(`spec:1903-1975`). Its input was the threat model and the diff, not the
plan's task decomposition. Verdict: do not merge as-is; fold six, then
the shell earns the merge. Twenty-four findings: one blocker, five
majors, eighteen minors. All six of the blocker and the majors folded,
plus nine of the eighteen minors; the remaining nine minors are
residuals, each carried to the rung that owns its closure.

This is the do-not-re-litigate record for passes 5b, 5c, and 5d. A
later pass reopening a folded item is redoing work the fold already
did; a later pass reopening a residual without new evidence is
reopening a ruling rather than closing a gap.

Numbering follows the close's own record: the nine residuals carry the
numbers the pass-close conductor named for them; the fifteen folded
items fill the remaining positions in the order the fold commits
(`0d3c644`, `307001b`) address them, since the review itself is not
committed to the tree.

## Findings

**1. Blocker: the refusal log wrote the full attacker-controlled
path.** `internal/serve/serve.go`'s refusal record wrote `r.URL.Path`
in full. A path refused by C3's length bound (`414`, bounded only by
`MaxHeaderBytes`) let an attacker write up to about 1 MiB to the
logger per request, a flood a single connection could sustain
indefinitely. Folded at `307001b`: the record now writes a 256-byte
prefix, a truncation marker, and the full length instead of the raw
path. `TestRefusalLogBoundsThePath` (`internal/serve/serve_test.go:741`)
proves the bound.

**2. Major: the session-cookie removal did not match `net/http`'s own
cookie-name parsing.** `withoutCookie` trimmed the whole `name=value`
segment before finding the `=`, where `net/http`'s `readCookies` trims
the name after the cut. A cookie sent as `sess =leak` was therefore a
cookie `r.Cookie` still answered to, yet the removal's own parity check
let it survive. Folded at `307001b`, cited as "cookie-name parity with
net/http's readCookies". `TestSessionCookieRemovedWhenTheNameCarriesSpace`
and `TestSessionCookieSpacedNameLeavesTheOthersReadable`
(`internal/serve/serve_test.go:791`, `:831`) prove the corrected parity.

**3. Major: a client's `Connection` header could delete a trusted
proxy's own forwarded headers.** Group 3's `Connection`-listed removal
ran without reference to group 2's carve-out, so `Connection:
X-Forwarded-For` from a trusted peer deleted the very header the
forwarded-address exemption exists to trust, collapsing `ClientIP` to
the proxy's own address. Folded at `307001b`: group 2's removal set is
now derived at `NewStrip` as the fixed family plus the configured
trusted-proxy header, and group 3 keeps those names standing for a
trusted peer. `TestStripWrapTrustedPeerKeepsForwardedListedInConnection`
(`internal/serve/strip_test.go:581`) proves the interaction.

**4. Major: the configured trusted-proxy header sat outside group 2's
removal set.** A client-address header an adopter configures at
runtime was not itself covered by the group-2 removal, so an
unallowlisted peer could set it directly. Folded at `307001b`, joined
with finding 3's fix (group 2's removal set now includes the
configured header). `TestStripWrapRemovesTheConfiguredProxyHeader`
(`internal/serve/strip_test.go:539`) proves the removal.

**5. Major: a smuggled trailer identity header survived a handler that
closed the body without reading it.** The trailer scrub read
`r.Trailer` at end-of-stream, which only happens once the body is
consumed; a handler that called `Body.Close()` without reading left the
announced trailer's smuggled value standing. Folded at `307001b`: the
trailer scrub gains a `Close` path. `TestStripWrapScrubsTrailerOnCloseWithoutRead`
(`internal/serve/strip_test.go:669`) proves the close-path scrub.

**6. Major: no signal context meant the graceful drain was unreachable
in production.** `Run`'s context-driven shutdown (`internal/serve/server.go`)
is real, but nothing in `main` ever canceled a context on `SIGINT` or
`SIGTERM`, so the drain path this pass built never fired outside a
test. Folded at `307001b`: `main` wires `signal.NotifyContext` and
`ExecuteContext`. `TestRunDrainsInFlightRequestAfterCtxDone`
(`internal/serve/server_test.go:308`) proves the drain mechanism the
wiring now reaches.

**7. Residual, owned by 5b: an absolute-form request target sets
`r.Host` to the client's chosen authority.** C3 rule 6 admits an
absolute-form target rather than refusing it, since refusing it would
break an adopter's own proxy; the residual is that nothing downstream
may derive an origin, a base URL, or a scheme from `r.Host`. 5b builds
every URL from the configured public origin, never `r.Host`, as an
assertion in its own test suite.

**8. Residual, no further rung: `identity_spelling_present` proves
order, not value.** The refusal-record assertion that identity
spellings are already gone by the time a refusal fires reads header-key
presence rather than header-value content, because Go populates an
announced trailer key with a nil value before the handler reads
anything. The assertion is sound for what C1's ordering claim needs; it
is not a value-content proof, and no later rung is asked to widen it.

**9. Folded minor: the wildcard bind address was accepted.**
`SplitBindAddrs` (`internal/serve/bindaddrs.go`) did not refuse `0.0.0.0`
or `::`, which C8's own loopback-default reasoning argues against:
the login surface is anonymous-reachable and the box is production.
Folded at `307001b`. `TestSplitBindAddrs`
(`internal/serve/bindaddrs_test.go:9`) carries the refusal rows.

**10. Folded minor: `Register` took no subtree-name validation.** A
subtree name that widened the owned mount past a single path segment
was accepted at construction rather than refused, echoing the
`mountBasePath` defect class C4 already guards against for prefix
matching. Folded at `307001b`. `TestRegisterRefusesNamesThatWidenTheMount`
and `TestRegisterAdmitsAnOrdinaryName`
(`internal/serve/serve_test.go:862`, `:904`) prove the validation.

**11. Folded minor: `validatePrefix`'s invalid shapes shared one
reason.** Every invalid prefix shape returned the same error text,
which makes a misconfiguration harder to diagnose than the registry's
own per-reason convention. Folded at `307001b`, "report each
invalid-prefix shape by its own reason".
`TestNewPrefixReportsEachShapeByItsOwnReason`
(`internal/serve/prefix_test.go:84`) proves the per-shape reasons.

**12. Folded minor: `TRACE` and `TRACK` reached routing.** Neither
method was refused ahead of the mux, so `TRACE` (which echoes the
request, headers included, back to the caller) and `TRACK` (IIS's
undocumented synonym) reached a registered subtree or the fallthrough
rather than being refused outright. Folded at `307001b`: both are
refused with `405` and no `Allow` header, a stated deviation from
`405`'s usual contract. `TestEchoMethodsRefusedBeforeRouting`
(`internal/serve/serve_test.go:922`) proves the refusal.

**13. Folded minor: `http.Server.ErrorLog` was unset.** With the
zero-value `ErrorLog`, `net/http`'s own protocol errors (a malformed
request line, a TLS handshake failure, a handler panic) went to
`log.Default`'s stderr rather than the shell's structured sink,
invisible to an operator watching the sink alone. Folded at `307001b`:
`ErrorLog` is wired to the shell's logger at `warn`.
`TestNewHTTPServerErrorLogReachesTheSlogSink`
(`internal/serve/server_test.go:425`) proves the wiring.

**14. Residual, accepted as is: no connection cap alongside the
120-second idle timeout.** The deadline policy (C6) bounds a single
slow connection; it does not bound how many idle connections a client
can hold open at once. No later rung is currently asked to add a
connection cap; the idle timeout is the stated mitigation.

**15. Folded minor: the strip's extras list needed a citation
completeness pass.** The review's second round asked for additional
trusted-header-SSO products (Grafana, Gitea, authentik, Tailscale) and
for a second look at candidates the first round dropped. Folded at
`307001b`: `X-Webauth-User`, the `X-Authentik-*` family,
`Tailscale-User-Login`, `Tailscale-User-Name`, and `X-Original-Url`
joined `identityExtras`; `Cf-Access-Authenticated-User-Email`, dropped
by the first round as uncitable, is retracted from the dropped list and
cited against Cloudflare's own Access tutorial. `X-Client-Ip`,
`X-Cluster-Client-Ip`, `X-Original-Method`, and `X-Rewrite-Url` stay
dropped: no page owned by a project that sets any of them was found.
`internal/serve/strip.go:26-118` carries the citation comment.

**16. Residual, owned by 5c: no `Referrer-Policy` header at 5a.** The
served-header set (`nosniff`, `Referrer-Policy: no-referrer`,
Content-Security-Policy) is 5c's deliverable against `spec:1386-1447`'s
page work; 5a's shell sets no response headers of its own beyond what
`net/http` sets by default.

**17. Residual, owned by rung 12: a `;`-separated path segment reaches
the fallthrough unrefused.** C3's rules do not name a semicolon as a
refused character. `http.ServeMux`'s own cleaning does not remove it
either, so a segment carrying `;matrix-param`-style syntax reaches
whichever handler owns the path, plain, for a router built to interpret
it differently to disagree with dubplate's own routing decision.

**18. Residual, owned by rung 12: a case-varying prefix reaches the
fallthrough unrefused.** C4's match is deliberately case-sensitive
(`/Prefix` is not `/prefix`), which is the stated policy; the residual
is that a case-varying path is not itself refused, only routed to the
fallthrough as ordinary traffic, and rung 12's driver is the first
consumer positioned to notice a case-insensitive backend served through
it.

**19. Folded minor: the client-IP-asserting header family was
incomplete.** The first round's group 2 covered
`X-Forwarded-For`/`X-Real-Ip`/`Forwarded` and the `X-Forwarded-Host`/
`-Proto`/`-Prefix` triad (R-R) but not the CDN- and edge-specific
client-IP headers (`Cf-Connecting-Ip`, `True-Client-Ip`,
`Fastly-Client-Ip`, `X-Azure-Clientip`), each of which a real front
sets and each of which an unauthenticated client could otherwise spoof
directly. Folded at `307001b`: all four join `forwardedAddressHeaders`,
cited against Cloudflare's, Fastly's, and Azure's own header
documentation. `TestStripWrapRemovesTheClientIPFamily`
(`internal/serve/strip_test.go:512`) proves the removal.

**20. Folded minor: `IP_FREEBIND` broke a non-Linux cross-compile.**
`freebind_linux.go`'s `syscall.IP_FREEBIND` reference had no build
constraint separating it from a non-Linux target build. Folded at
`307001b`: `freebind_linux.go` carries `//go:build linux`, and
`freebind_other.go` supplies the no-op control hook for every other
target.

**21. Folded minor: two client-IP candidates and two identity-header
candidates needed a second citation pass.** `X-Client-Ip` and
`X-Cluster-Client-Ip` are widely quoted in client-IP-extraction
libraries with no citable front actually setting either; `X-Original-Method`
and `X-Rewrite-Url` are widely quoted in third-party `auth_request`
material with no citable owning project. Folded at `307001b` as a
disposition, not an addition: all four stay out of their respective
lists, each with the reason recorded beside it, so a later rung
re-considering one starts from a documented "not yet citable" rather
than rediscovering the absence.

**22. Residual, owned by 5d: the trusted-proxy address set is a plain
string comparison rather than `netip`-typed.** R27's derivation
compares the peer address as a string against the configured proxy
address set; a `netip.Addr`/`netip.Prefix` typed comparison would admit
a CIDR range and normalize representation (`::ffff:127.0.0.1` against
`127.0.0.1`) where the string form cannot. 5d's trusted-proxy
integration class is where the configured set gains real shape,
against which the typed comparison decision belongs.

**23. Residual, owned by 5d: `check` and `doctor` do not validate a
serve value.** `dubplate-server check` and `dubplate-server doctor`
report the configured prefix, bind addresses, and port as whatever
`Load` accepted; neither validates that the bind address parses to a
real, bindable address or that the prefix is reachable. 5d's `doctor`
row (R-G) is where this reporting gains a live-evidence check rather
than a configuration echo.

**24. Residual, owned by rung 15 via the wall's own contract: a
registered threshold is accepted by `Load` with any value.** The
rejectNamespace wall's fifth classification (a registered threshold is
a declared name) makes `Load` accept a threshold-classified name
regardless of the value supplied; no per-threshold value validation
runs at load time beyond the registry's own default-presence check.
This is a property of the rung 1b wall contract rather than something
this pass could close, and is recorded here so a later audit does not
mistake it for a rung 5a gap.

## The verified-sound table

The subset of the plan's threat-model rubric each landed test proves,
read against the claims section (`plan:1160-1266`):

| Rubric threat | Proving test |
| --- | --- |
| Header spoofing, both spellings, before and after the strip | `TestStripWrapRemovesEveryIdentitySpelling` (`strip_test.go:85`), `TestUnderscoreAndDashAreDistinctCanonicalKeys` (`strip_test.go:32`) |
| Trailer-borne identity, announced and unannounced, read before and after the body | `TestStripWrapClearsAnnouncedTrailerOnceBodyIsRead` (`strip_test.go:103`), `TestStripWrapClearsUnannouncedTrailerOnceBodyIsRead` (`strip_test.go:406`), `TestStripWrapScrubsTrailerOnCloseWithoutRead` (`strip_test.go:669`) |
| Header deletion via `Connection`, including the keep-set and the websocket carve-out | `TestStripWrapConnectionListedRemoval` (`strip_test.go:155`), `TestStripWrapConnectionStillRemovesAnOrdinaryListedHeader` (`strip_test.go:644`), `TestStripWrapExemptedHeaderSurvivesConnectionListing` (`strip_test.go:467`) |
| The forwarded family, allowlisted and non-allowlisted peers | `TestStripWrapGroup2ForwardedAddress` (`strip_test.go:248`), `TestStripWrapTrustedPeerKeepsForwardedListedInConnection` (`strip_test.go:581`), `TestStripWrapRemovesTheClientIPFamily` (`strip_test.go:512`) |
| Encoded-slash and dot-segment bypasses | `TestNormalize` (`normalize_test.go:50`), `TestUnencodedPathCarriesEmptyRawPath` (`normalize_test.go:110`) |
| Case games at the prefix boundary | `TestRoute` (`prefix_test.go:124`) |
| Oversized paths | `TestNormalize`'s length-bound rows (`normalize_test.go:50`) |
| Slowloris and slow-body resistance | Not test-proven; rests on the timeout constants (`TestNewHTTPServerTimeouts`, `server_test.go:23`), per declared coverage gap 5 |
| Method games on the fallthrough and the terminator | `TestEchoMethodsRefusedBeforeRouting` (`serve_test.go:922`), `TestUnclaimedOwnedPathReachesTheTerminator` (`serve_test.go:559`), `TestEmptyFallthroughAnswers404` (`serve_test.go:261`) |
| A response outliving the default write deadline | `TestWriteDeadlineExtension` (`server_test.go:69`) |
| Refusal-log flood via an oversized path | `TestRefusalLogBoundsThePath` (`serve_test.go:741`) |
