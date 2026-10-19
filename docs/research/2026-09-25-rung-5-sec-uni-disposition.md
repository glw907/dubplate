# Rung 5: the SEC and UNI disposition table

This table maps each of the 50 findings in the two login-spec adversarial reviews to the clause
that absorbed it, the code and test that prove it, and its status. The two reviews are the
security review (`docs/research/2026-09-02-login-spec-reviews/security.md`, SEC) and
the universality review (`docs/research/2026-09-02-login-spec-reviews/universality.md`,
UNI). The build spec makes both texts birth requirements, and the rung 5 spec names this table as
5d's deliverable and the adversarial cycle's input (spec:2003-2009).

Every code and test pointer was read at `03e1518` on master with `grep -n`, never recalled. A
pointer names the line where the cited declaration or statement starts.

## Sources and abbreviations

| Short form | Document |
| --- | --- |
| SEC, UNI | The two review texts above, cited by line (`SEC:11` is the line a finding starts on) |
| spec | `docs/specs/2026-09-08-rung-5-identity-auth.md`, the rung 5 spec |
| BS | `docs/specs/2026-09-02-greenfield-server-build.md`, the build spec; its auth birth requirements start at BS:479 |
| plan | `docs/plans/2026-09-25-rung-5d-boundary-deployment-close.md`, revision 2.2 |
| HISTORY, STATUS | `docs/HISTORY.md` and `docs/STATUS.md` at `03e1518` |
| arch reads | `docs/research/2026-10-05-rung-5d-i-architecture-reads.md` |
| fold record | `docs/research/2026-09-25-rung-5d-plan-fold.md` |

Code paths are repository-relative. A test cell gives the test's name and the line its `func`
declaration starts on.

## How to read the table

Each finding gets an ID from its severity and its order in its review text:

- SEC: `SEC-B1` to `SEC-B6` (blockers), `SEC-M1` to `SEC-M13` (majors), `SEC-m1` to `SEC-m12`
  (minors).
- UNI: `UNI-C1` to `UNI-C3` (criticals), `UNI-H1` to `UNI-H6` (highs), `UNI-M1` to `UNI-M7`
  (mediums), `UNI-L1` to `UNI-L3` (lows).

The status column takes one of three values:

- **Absorbed.** A clause or ruling took the finding, and code plus a test prove it at HEAD. Where
  no test can reach the property, the row says "by design" and cites the clause or ruling that
  ratified the design.
- **Absorbed, residual.** The finding is absorbed, and a stated part of it remains. The row names
  the residual and where it is recorded.
- **Carried.** The finding belongs to a later rung. The row names the rung and cites the record
  that carries it.

## SEC blockers (6)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| SEC-B1 (SEC:11) | The `/rest/*` and `/share/*` bypasses forward a client-supplied identity header to the streamer, which is account takeover. | BS:491-502 (strip on ingress, before routing, for every path, bypasses included; inject only on the session-gated path); spec:790-794; spec:933-934 (the strip is the outermost routing step). | `Shell.Handler` composes the strip around normalization and routing, `internal/serve/serve.go:176`; `Strip.Wrap`, `internal/serve/strip.go:333`; the production shell passes the zero front exemption, `cmd/dubplate-server/serve.go:259`. Tests: `TestOrderStripBeforeFallthrough` (`internal/serve/serve_test.go:94`), `TestOrderStripBeforeOwnedSubtree` (`internal/serve/serve_test.go:69`), `TestNewServeShell_StripCarriesNoFrontExemption` (`cmd/dubplate-server/serve_test.go:1895`), `TestDoctorBoundaryProbe_EndToEnd` (`cmd/dubplate-server/doctorprobe_test.go:385`). | Absorbed, residual. Injection on the proxy path is rung 12's (spec:1782-1787, spec:1855-1860). Navidrome's own `/share/*` links stay live until 0.8, an owner-accepted residual (BS:68-78). `FrontExemption` carries a latent dual-stack peer-match bug behind no production caller, carried to rung 12 (STATUS:56; arch reads:12-16). |
| SEC-B2 (SEC:14) | Pinning the streamer's trusted sources to the compose bridge subnet trusts every container on it, the tunnel container included. | BS:498-500 (the trusted source is serve's single observed address, never a bridge subnet, and no other network path reaches the streamer); spec:839-842. | None at HEAD: rung 5 builds no driver and no externalized-auth configuration (spec:1784-1787). | Carried to rung 12, which owns the driver's trusted-source pin (spec:839-842, spec:1784-1787). |
| SEC-B3 (SEC:17) | The live tunnel rule sends the music hostname straight to Navidrome, around any SSO proxy. | BS:500-502 (the ingress flip precedes enabling externalized auth, stated wherever the cutover is planned); spec:839-842. | None at HEAD: `config/cloudflared-tunnel-config.json` still routes to Navidrome only, and Navidrome runs no externalized auth. The first deploy installs the serve unit disabled, and nothing routes to it (plan:145). | Carried to rung 16, which runs the ingress flip before enabling externalized auth (spec:1866-1868). |
| SEC-B4 (SEC:20) | `/api/*` collides with the streamer's own API, so a mux miss would decide the auth boundary. | BS:516-520 (one owned prefix; configuration decides the boundary); spec:905-922. | `Prefix.Route`, `internal/serve/prefix.go:149`; `Prefix.Owns`, `internal/serve/prefix.go:116`; `Terminator`, `internal/serve/prefix.go:131`. Tests: `TestRoute` (`internal/serve/prefix_test.go:124`), `TestOrderPrefixRequiresSegmentBoundary` (`internal/serve/serve_test.go:160`), `TestUnclaimedOwnedPathReachesTheTerminator` (`internal/serve/serve_test.go:559`), `TestNewServeShell_MountsTheThreeAuthRoutes` (`cmd/dubplate-server/serve_test.go:391`). | Absorbed. |
| SEC-B5 (SEC:23) | A per-`RemoteAddr` limit is one global key behind the tunnel, and a client-address header is spoofable unless only a configured proxy may set it. | BS:566-571; R22 (spec:572-577); R27 (spec:611-622); spec:844-870. | `ClientIP` reads the header only from a configured proxy, `internal/serve/clientaddr.go:171`; trust decisions read `PeerAddress`, `internal/serve/clientaddr.go:131`; the resolver, `cmd/dubplate-server/proxy.go:52`; the shell keys the limiters on `ClientIP`, `cmd/dubplate-server/serve.go:297`; doctor's client ip row, `cmd/dubplate-server/doctorprobe.go:50`. Tests: `TestClientIP` (`internal/serve/clientaddr_test.go:74`), `TestResolveTrustedProxy` (`cmd/dubplate-server/proxy_test.go:24`), `TestNewServeShell_TrustedProxyKeysTheLinkLimiter` (`cmd/dubplate-server/serve_test.go:1815`), `TestRequestLink_ClientAddressIsTheShells` (`internal/auth/handlers_test.go:869`), `TestDoctorBoundaryRows_Perturbations` (`cmd/dubplate-server/doctorprobe_test.go:86`). | Absorbed, residual. The estate's template sets the mode to direct until the cutover pins the tunnel container's address (plan:567-585). Every group 2 name survives from a trusted peer, the configured header and the rest alike, carried to rung 12 (plan:1223-1224; fold record:153). |
| SEC-B6 (SEC:26) | Consuming the link on GET invites scanner prefetch, the token in logs, and login CSRF. | BS:521-527; R32 (spec:671-688); spec:1051-1082; the B4 erratum (spec:2048-2059). | The link GET serves the client-rendered interstitial `no-store`, `cmd/dubplate-server/web.go:107`; consumption is POST-only, `cmd/dubplate-server/serve.go:304`; its log names the path and never the query, `internal/auth/handlers.go:337`. Tests: `TestConsumeLink_SignsIn` (`internal/auth/handlers_test.go:333`), `TestNewServeShell_MintedLinkNeverLogsItsToken` (`cmd/dubplate-server/serve_test.go:1724`), `TestLogsCarryNoCredential` (`internal/auth/handlers_test.go:923`), `TestOriginCheckCoversSubtreeAndFallthrough` (`internal/serve/serve_test.go:983`). | Absorbed, residual. The link URL carries the token to a mail scanner and browser history, bounded by single use and a short expiry (R32, spec:671-688; threat model, spec:1990-1994). |

## SEC majors (13)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| SEC-M1 (SEC:33) | Prefix matching on a decoded path lets `%2F` and `..` smuggle a path through a bypass prefix. | BS:600-611 (normalization before prefix routing), landed at rung 5 by the B12 erratum (spec:2038-2046); spec:935-944; the rule-set erratum (spec:2132-2153). | `Normalize`, `internal/serve/normalize.go:83`; `pathRefusal`, `internal/serve/normalize.go:98`; normalization wraps routing, `internal/serve/serve.go:175`. Tests: `TestNormalize` (`internal/serve/normalize_test.go:50`), `TestOrderRawVsDecodedRefusedBeforeAnyHandler` (`internal/serve/serve_test.go:123`), `TestRefusalLogsIdentityAlreadyGone` (`internal/serve/serve_test.go:184`). | Absorbed, residual. The blanket ban on percent-encoding in a path is revisited at rung 12 against the streamer's real client paths (spec:2139-2142). |
| SEC-M2 (SEC:36) | The salvage origin check sat inside the `/upload` mount and covered no new surface. | BS:560-564 (one CSRF rule for every cookie-carrying, state-changing surface); R15 (spec:494-525); spec:1212-1236. | The origin check wraps both halves of prefix routing at the root, `internal/serve/serve.go:173`; `originCheck.allowed`, `internal/serve/origin.go:103`. Tests: `TestOriginCheck` (`internal/serve/origin_test.go:29`), `TestOriginCheckCoversSubtreeAndFallthrough` (`internal/serve/serve_test.go:983`), `TestNewServeShell_StateChangingFallthroughRequestIsDenied` (`cmd/dubplate-server/serve_test.go:545`), `TestNewServeShell_RosterPOSTWithoutOriginSignalIsDenied` (`cmd/dubplate-server/serve_test.go:805`). | Absorbed. |
| SEC-M3 (SEC:39) | `ReverseProxy` hygiene is unspecified: `Director` appends to a client's `X-Forwarded-For`, and `Connection: Remote-User` deletes the injected identity. | BS:600-611; the B12 erratum (spec:2038-2046); spec:816-831; the group 2 erratum (spec:2124-2130). | Group 2 lists `X-Real-Ip`, `Forwarded`, and `X-Forwarded-Host`, `internal/serve/strip.go:174`; group 1 lists `X-Forwarded-User`, `internal/serve/strip.go:112`; `stripForwardedAddress`, `internal/serve/strip.go:460`; `stripConnectionListed`, `internal/serve/strip.go:496`. Tests: `TestStripWrapGroup2ForwardedAddress` (`internal/serve/strip_test.go:248`), `TestStripWrapRemovesTheClientIPFamily` (`internal/serve/strip_test.go:512`), `TestStripWrapConnectionListedRemoval` (`internal/serve/strip_test.go:155`). | Absorbed, residual. `Rewrite` with `SetXForwarded` and injection inside `Rewrite` are rung 12's (spec:2038-2046). Group 2 survives from a trusted peer, carried to rung 12 (plan:1223-1224; fold record:153). |
| SEC-M4 (SEC:42) | The strip set is not enumerated, and the underscore spelling escapes a dash-form delete. | BS:491-496 (driver header plus a fixed extras list, underscore variants included); spec:796-814. | `identityExtras`, `internal/serve/strip.go:112`; `identitySpellings` derives each underscore twin, `internal/serve/strip.go:210`; the driver's header is a construction value, empty at rung 5, `cmd/dubplate-server/serve.go:259`. Tests: `TestUnderscoreAndDashAreDistinctCanonicalKeys` (`internal/serve/strip_test.go:32`), `TestStripWrapRemovesEveryIdentitySpelling` (`internal/serve/strip_test.go:85`), `TestProbeRoute_ThroughShellCountsNothing` (`internal/serve/probe_test.go:64`). | Absorbed, residual. The list admits a name only with a citation (`internal/serve/strip.go:11-24`). SEC also names `X-Remote-User` and `X-Authenticated-User`; neither is in the list, and its admission record names neither as admitted or dropped (`internal/serve/strip.go:26-111`). This table records that gap (residue item 6). The driver's own header joins at rung 12 (spec:796-801). |
| SEC-M5 (SEC:45) | The link's host is unspecified, so a forged `Host` header poisons a signed link. | BS:552-559 (the public URL is required when magic-link is active; link hosts never derive from request headers); spec:1048-1049; the public-URL erratum (spec:2463-2468). | `LinkURL` builds from the configured origin and prefix, `internal/auth/token.go:172`; `validatePublicURL`, `cmd/dubplate-server/auth.go:441`; the source's required set, `cmd/dubplate-server/auth.go:43`. Tests: `TestLinkURL` (`internal/auth/token_test.go:390`), `TestValidateMagicLinkSource_PublicURL` (`cmd/dubplate-server/auth_test.go:265`), `TestRegisterAuthConfig_MagicLinkSource` (`cmd/dubplate-server/auth_test.go:40`), `TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree` (`cmd/dubplate-server/serve_test.go:628`). | Absorbed. |
| SEC-M6 (SEC:48) | Sessions have no absolute lifetime, no per-person revocation, and no offboarding path. | BS:537-544; fork 3 (spec:220-233); spec:1095-1105; R30 (spec:647-658); the anchor erratum (spec:2267-2275). | `Session.clampExpiry`, `internal/auth/session.go:147`; `VerifySession` refuses past the cap, `internal/auth/session.go:219`; `EpochStore.CheckSession` reads the anchor, `internal/auth/epoch.go:225`; `EpochStore.CheckLink`, `internal/auth/epoch.go:253`. Tests: `TestRefreshedSessionRefusedAtTheAbsoluteCap` (`internal/auth/session_test.go:348`), `TestCheckSessionRefreshedAfterBumpStillRefused` (`internal/auth/epoch_test.go:110`), `TestCheckLink` (`internal/auth/epoch_test.go:134`), `TestBumpReadDurable` (`internal/auth/epoch_test.go:31`). | Absorbed, residual. `EpochStore.Bump` (`internal/auth/epoch.go:103`) has no production caller, so no operator can revoke until the revoke verb lands (STATUS:58-59; the 6d plan, `docs/plans/2026-09-26-rung-6d-identity-serve-store.md:93-109`). The runbook's streamer-deprovision step (spec:1100-1105) appears in no task's outline (residue item 7). |
| SEC-M7 (SEC:51) | The two HMAC secrets need domain separation, a length floor, a rotation overlap, and unambiguous serialization. | BS:552-559; spec:1029-1046; the serialization erratum (spec:2285-2291). | `secretLengthFloor`, `cmd/dubplate-server/auth.go:71`; `linkPurpose`, `internal/auth/token.go:36`; `sessionPurpose`, `internal/auth/session.go:35`; length-prefixed MAC input, `internal/auth/token.go:324`; `VerifyLink` takes the previous secret, `internal/auth/token.go:143`. Tests: `TestValidateMagicLinkSource_SecretLengthFloor` (`cmd/dubplate-server/auth_test.go:228`), `TestSessionAndLinkCannotCross` (`internal/auth/session_test.go:446`), `TestMACLengthPrefixesResistFieldShifting` (`internal/auth/token_test.go:247`), `TestVerifyLinkAcceptsPreviousSecret` (`internal/auth/token_test.go:117`). | Absorbed, residual. Both HMAC keys and the SMTP password sit in every `music`-uid verb's environment, the shadow's included, carried to rung 16 (plan:140, plan:350). |
| SEC-M8 (SEC:54) | A memory-only nonce set forgives replay across a restart, and the insert order is unstated. | BS:521-526; R11 (spec:430-452); R31 (spec:660-669); spec:1039-1041. | The create-exclusive insert, `internal/auth/nonce.go:155`; the consume order is MAC, roster, epoch, insert, `internal/auth/handlers.go:361`; the sweep driver, `cmd/dubplate-server/serve.go:432`. Tests: `TestInsertRefusesReplayAcrossRestart` (`internal/auth/nonce_test.go:105`), `TestFailedMACNeverReachesNonceStore` (`internal/auth/nonce_test.go:125`), `TestConcurrentInsertAdmitsOne` (`internal/auth/nonce_test.go:542`), `TestValidateMagicLinkSource_NonceRetention` (`cmd/dubplate-server/auth_test.go:346`). | Absorbed. |
| SEC-M9 (SEC:57) | Unauthenticated input reaches unbounded maps; bodies, address length, and address case are uncapped. | BS:566-575; R16 (spec:527-533); spec:1182-1184; spec:1202-1206. | `maxLimiterKeys`, `internal/auth/limit.go:33`; `readForm` caps the body, `internal/auth/handlers.go:396`; `NormalizeAddress`, `internal/identity/identity.go:50`, through `mail.ParseAddress`, `internal/identity/identity.go:134`, under `MaxAddressLength`, `internal/identity/identity.go:23`. Tests: `TestKeyedLimiterBoundHoldsUnderFlood` (`internal/auth/limit_test.go:147`), `TestOversizeBodyRefusedBeforeParse` (`internal/auth/handlers_test.go:530`), `TestNormalizeAddress` (`internal/identity/identity_test.go:141`), `TestRoster_LookupCaseInsensitive` (`internal/auth/roster_test.go:380`). | Absorbed, residual. Past capacity the per-address map admits a new key untracked, unreachable below 10,000 members (HISTORY:3412-3415). |
| SEC-M10 (SEC:60) | A process-wide 30-second write timeout cuts every SSE and audio stream through the proxy. | BS:600-611 (per-route write-timeout handling); spec:986-1001; the B12 erratum (spec:2038-2046); the read-timeout erratum (spec:2155-2163). | `newHTTPServer`, `internal/serve/server.go:92`, with `defaultWriteTimeout` at `internal/serve/server.go:69`; a handler extends its own deadline through `http.ResponseController`. Tests: `TestNewHTTPServerTimeouts` (`internal/serve/server_test.go:23`), `TestWriteDeadlineExtension` (`internal/serve/server_test.go:69`). | Absorbed, residual. The streaming routes that clear the deadline, and their slow-client bound, are rung 12's (spec:986-993, spec:1855-1860). |
| SEC-M11 (SEC:63) | The session cookie reaches the streamer and lacks a `__Host-` prefix. | BS:529-531; spec:833-837; spec:1086-1093; the R-E erratum (spec:2093-2099). | `SessionCookieName`, `internal/auth/session.go:20`; `SessionCookie`, `internal/auth/session.go:172`; the fallthrough removal, `internal/serve/serve.go:242`. Tests: `TestSessionCookieName` (`internal/auth/session_test.go:43`), `TestSessionCookieAttributes` (`internal/auth/session_test.go:54`), `TestSessionCookieRemovedAtFallthroughPresentAtOwnedPrefix` (`internal/serve/serve_test.go:307`), `TestNewServeShell_RemovesTheSessionCookieAtTheFallthrough` (`cmd/dubplate-server/serve_test.go:344`). | Absorbed. |
| SEC-M12 (SEC:66) | The identity header carries the roster's free-text name, with no uniqueness or charset rule. | BS:579-586; R3 (spec:326-335); spec:1118-1166; the write-path erratum (spec:2410-2419). | `Member` keeps the principal apart from the display name, `internal/auth/roster.go:163`; `memberRecord.validate`, `internal/auth/roster.go:575`; `streamerUsernamePattern`, `internal/auth/roster.go:52`; `principalCollision`, `internal/auth/roster.go:830`. Tests: `TestOpenRoster_ValidationRows` (`internal/auth/roster_test.go:131`), `TestRoster_EffectivePrincipalUniqueness` (`internal/auth/roster_test.go:1876`), `TestValidateWriteCandidate_StreamerUsernameShape` (`internal/auth/roster_test.go:459`), `TestAddMember_ConcurrentStreamerUsernameCollision` (`internal/auth/roster_test.go:1125`), `TestDoctorCmd_RosterRow` (`cmd/dubplate-server/doctor_test.go:813`). | Absorbed, residual. The header injection is rung 12's. Rung 12 re-validates every existing row's principal (HISTORY:2373-2375) and must refuse a non-admin principal that maps onto a streamer admin (HISTORY:3416-3419). |
| SEC-M13 (SEC:69) | Two identity sources, one of them unconfigured, must never both be consulted. | BS:587-591 (exactly one active source, selected by config); R5 (spec:349-361); R29 (spec:633-645); spec:1645-1653. | The selector wall row, `cmd/dubplate-server/auth.go:21`; the Access keys are superseded, `internal/config/keys.go:133`; no Access verifier exists, and the strip removes `Cf-Access-Jwt-Assertion`, `internal/serve/strip.go:124`. Tests: `TestRegisterAuthConfig_SelectorIsOwnedNow` (`cmd/dubplate-server/auth_test.go:20`), `TestLoad_IdentitySource` (`internal/config/config_test.go:735`), `TestDoctorCmd_UnsetSelectorExitsNonzero` (`cmd/dubplate-server/auth_test.go:100`), `TestNewServeShell_StripCarriesNoFrontExemption` (`cmd/dubplate-server/serve_test.go:1895`). | Absorbed, residual. The front exemption, the mechanism a later front uses, matches peers by raw string and misses an IPv4 front on a dual-stack listener, carried to rung 12 (STATUS:56; arch reads:12-16). |

## SEC minors (12)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| SEC-m1 (SEC:76) | Redirect validation needs its bypass forms spelled out. | R33 (spec:690-702); spec:1268-1286; the R33 erratum (spec:2366-2375). | Consumption reads only the token field, `internal/auth/handlers.go:348`, and answers JSON with no destination, `internal/auth/handlers.go:356`. Test: `TestConsumeLink_SignsIn` (`internal/auth/handlers_test.go:333`). | Absorbed by design at rung 5: no caller-supplied destination exists (R33). The validator is carried to rung 12 with its destination parameter (spec:2366-2375). |
| SEC-m2 (SEC:79) | The kind answer for an unknown address prices neither the broken limit key nor the published admin contact. | Fork 4 (spec:235-251); R25 (spec:597-599); R28 (spec:624-631). | `unknownAddressAnswerThreshold`, `cmd/dubplate-server/auth.go:210`; the generic answer, `internal/auth/handlers.go:237`; the global mail ceiling, `internal/auth/limit.go:109`. Tests: `TestRequestLink_GenericAnswersAlike` (`internal/auth/handlers_test.go:600`), `TestRequestLink_GenericOverLimitAnswersAsUnknown` (`internal/auth/handlers_test.go:1128`), `TestRequestLink_ClientFloodSendsNoFurtherMail` (`internal/auth/handlers_test.go:842`), `TestLinkMailCeilingIgnoresKeys` (`internal/auth/limit_test.go:118`), `TestValidateMagicLinkSource_UnknownAddressAnswer` (`cmd/dubplate-server/auth_test.go:709`). | Absorbed, residual. The kind answer reveals membership and the first admin's contact (threat model, spec:1994), and the runbook names it (plan:750-752). |
| SEC-m3 (SEC:82) | Distinct consumption errors tell an attacker whether a held token was real. | BS:525-526; spec:1058-1059. | Every refusal after the limit answers one outcome, `internal/auth/handlers.go:351`, and the cause reaches the log alone. Test: `TestConsumeLink_Refusals` (`internal/auth/handlers_test.go:392`). | Absorbed. |
| SEC-m4 (SEC:85) | Security headers and edge caching are unspecified for proxied responses. | R32 (spec:671-688); the served-header erratum (spec:2497-2504); BS:600-611 (identity-injected responses are `Cache-Control: private` and never edge-cached). | `securityHeaders`, `internal/serve/headers.go:34`, wraps the whole chain, `internal/serve/serve.go:177`. Tests: `TestShellHandler_SecurityHeadersOnEveryResponse` (`internal/serve/headers_test.go:18`), `TestNewServeShell_HeaderStepOnEveryResponse` (`cmd/dubplate-server/serve_test.go:1226`). | Absorbed, residual. The proxied surface's headers and cache rule are rung 12's (spec:2038-2046). Rung 12 reconciles the header step with the streamer (HISTORY:2366-2367) and takes the shared-origin residual (HISTORY:2376-2377). |
| SEC-m5 (SEC:88) | The auth pages need `Referrer-Policy` and no third-party origin as requirements. | R32 (spec:671-688); O2 (spec:156-159); BS:521-527. | `securityHeaders` sets `Referrer-Policy: no-referrer` on every response, `internal/serve/headers.go:34`; gate G3 refuses an external origin in the source and the built bundle, `web/scripts/gates/g3.ts:1`, run from `Makefile:96`. Test: `TestShellHandler_SecurityHeadersOnEveryResponse` (`internal/serve/headers_test.go:18`). | Absorbed, residual. An old browser's `Referrer-Policy` gap goes to the 5d runbook (HISTORY:2379; plan:1187). |
| SEC-m6 (SEC:91) | The reverse-proxy auth setting's name may not match the pinned Navidrome version. | spec:1784-1787 (the driver's externalized-auth configuration is rung 12's). | None at HEAD: no driver and no externalized-auth setting exist. | Carried to rung 12 (spec:1784-1787). |
| SEC-m7 (SEC:94) | SMTP headers built from roster values need explicit CRLF rejection. | spec:1304-1310 (the seam owns the guard over every header-valued argument); BS:579-581. | `guardHeaderValue`, `internal/notify/notify.go:148`; roster addresses pass `mail.ParseAddress`, `internal/identity/identity.go:134`. Tests: `TestSend_HeaderInjectionGuard` (`internal/notify/notify_test.go:464`), `TestOpenRoster_ValidationRows` (`internal/auth/roster_test.go:131`). | Absorbed, residual. The SMTP transport is not exercised on the box, carried to rung 16 (plan:1458). |
| SEC-m8 (SEC:97) | A page echoing the submitted address must escape it. | O2 (spec:156-159): the pages are SvelteKit static and client-rendered, so no server template renders a submitted value. | The login page binds the address to the input's value only, `web/src/routes/login/+page.svelte:137`. Test: `TestNewWebRoutes_SiteNameMeta` (`cmd/dubplate-server/web_test.go:1681`) proves the one operator value served into the documents is escaped. | Absorbed by design (O2). |
| SEC-m9 (SEC:100) | "First user becomes admin" is documentation, not a control. | Forks 1 and 5 (spec:178-200, spec:253-262). | `member add --admin` is the only admin grant, `cmd/dubplate-server/member.go:63`; the web path refuses the grant, `internal/auth/roster.go:96`. Tests: `TestMemberAddCmd_NoAdminRefusedOnCreate` (`cmd/dubplate-server/member_test.go:95`), `TestRosterWrites_AdminInvariantAndGrant` (`internal/auth/roster_test.go:1670`). | Absorbed, residual. The streamer half (its own admin before SSO, and bring-up proving it) is rung 12's (spec:253-262; HISTORY:3416-3419). |
| SEC-m10 (SEC:103) | Retiring Access exposes the Go stack to anonymous traffic: bind narrowly, cap bodies, bound limiters, log every auth event, keep a WAF fallback, and record `/rest/*`. | BS:566-578; spec:1202-1210; spec:1599-1608 (loopback default, never `0.0.0.0`). | The bind default is loopback, `cmd/dubplate-server/serve.go:90`; `Listen` binds exact addresses, `internal/serve/server.go:113`; doctor's bind row, `cmd/dubplate-server/doctorprobe.go:125`; auth events log the client IP, `internal/auth/handlers.go:189`. Tests: `TestListenBindsExactAddresses` (`internal/serve/server_test.go:142`), `TestBindAddrsRowStatus` (`cmd/dubplate-server/doctorprobe_test.go:271`), `TestRequestLink_ClientAddressIsTheShells` (`internal/auth/handlers_test.go:869`), `TestOversizeBodyRefusedBeforeParse` (`internal/auth/handlers_test.go:530`). | Absorbed, residual. The shell logs identity-header presence only on its own refusals (`internal/serve/serve.go:427`), so a header-strip hit is never its own auth event (residue item 8). No spec, plan, or ledger records a WAF or Turnstile fallback (residue item 9). `/rest/*` is a threat-model residual (spec:1985-1987) with no path through serve before rung 12. |
| SEC-m11 (SEC:106) | `SameSite=Lax` still sends the cookie on a cross-site top-level GET, which the proxy would inject identity into. | BS:560-564; R15 (spec:494-525); BS:600-611 (the proxy refuses to inject on a non-idempotent request lacking a same-origin signal). | The origin check covers the fallthrough, `internal/serve/serve.go:173`. Tests: `TestSessionCookieAttributes` (`internal/auth/session_test.go:54`), `TestOriginCheckCoversSubtreeAndFallthrough` (`internal/serve/serve_test.go:983`), `TestNewServeShell_StateChangingFallthroughRequestIsDenied` (`cmd/dubplate-server/serve_test.go:545`). | Absorbed, residual. The injection refusal is rung 12's (BS:600-611; spec:1855-1860). |
| SEC-m12 (SEC:109) | Rolling refresh has no trigger boundary, so a cross-site navigation keeps an idle session alive. | BS:537-538; spec:1095-1098. | `MagicLinkSeam.VerifyIdentity` refreshes only when `sameOrigin` holds, `internal/auth/seam.go:102`; `RequireAdmin` derives it from `Sec-Fetch-Site`, `internal/auth/admin.go:112`. Tests: `TestMagicLinkSeam_VerifyIdentity_CrossOriginDoesNotRefresh` (`internal/auth/seam_test.go:219`), `TestRequireAdmin_RefreshOnlySameOrigin` (`internal/auth/admin_test.go:436`), `TestRefreshedSessionRefusedAtTheAbsoluteCap` (`internal/auth/session_test.go:348`). | Absorbed. |

## UNI criticals (3)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| UNI-C1 (UNI:7) | Routing rule 2 hands `/api/*` to dubplate, which collides with the streamer's own API. | BS:516-520; spec:905-922. | As SEC-B4: `internal/serve/prefix.go:149`, `internal/serve/prefix.go:116`, `internal/serve/prefix.go:131`. Tests: `TestRoute` (`internal/serve/prefix_test.go:124`), `TestOrderPrefixRequiresSegmentBoundary` (`internal/serve/serve_test.go:160`), `TestUnclaimedOwnedPathReachesTheTerminator` (`internal/serve/serve_test.go:559`). | Absorbed. |
| UNI-C2 (UNI:10) | The trusted-source pin trusts the whole bridge, the tunnel container included, and needs an ordering constraint. | BS:498-502; spec:839-842. | None at HEAD, as SEC-B2 and SEC-B3. | Carried: the pin to rung 12 (spec:1784-1787), the ingress flip before externalized auth to rung 16 (spec:1866-1868). |
| UNI-C3 (UNI:13) | One roster `name` is the inbox directory, the streamer account, and the auth principal at once. | BS:579-586; R3 (spec:326-335); spec:1161-1166. | `Member` carries the address, display name, and `streamer_username` apart, `internal/auth/roster.go:163`; `Member.principal`, `internal/auth/roster.go:817`. Tests: `TestIdentity_PrincipalDistinctFromAddress` (`internal/identity/identity_test.go:434`), `TestRoster_EffectivePrincipalUniqueness` (`internal/auth/roster_test.go:1876`), `TestDoctorCmd_RosterRow` (`cmd/dubplate-server/doctor_test.go:813`). | Absorbed, residual. A principal that maps a non-admin onto a streamer admin is refused at rung 12 (HISTORY:3416-3419). |

## UNI highs (6)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| UNI-H1 (UNI:16) | The design assumes one hostname and never scopes the cookie. | BS:529-537 (same-origin is a stated precondition of SSO; no cookie-domain widening); spec:1086-1093. | `SessionCookie` sets `Path=/` and no `Domain`, `internal/auth/session.go:172`. Test: `TestSessionCookieAttributes` (`internal/auth/session_test.go:54`). | Absorbed. |
| UNI-H2 (UNI:19) | The forward-auth shape, the universal deployment, is never designed. | R21 alone (spec:566-570), with BS:504-515 and spec:884-903. | None at HEAD: R-5d17 removed the verify endpoint from 5d (plan:608-632). The front exemption's latent dual-stack bug rides with it (STATUS:56; arch reads:12-16). | Carried to rung 12 (R-5d17, plan:608-632). |
| UNI-H3 (UNI:22) | Nothing carries cookie-derived provenance, so admin verbs cannot refuse it at the seam. | BS:544-548; R17 (spec:535-540); spec:1238-1262; the channel erratum (spec:2428-2436). | `identity.Channel`, `internal/identity/identity.go:151`; `Identity.Valid`, `internal/identity/identity.go:208`; the seam stamps the cookie channel, `internal/auth/seam.go:102`. Tests: `TestChannel_StringAndValid` (`internal/identity/identity_test.go:293`), `TestIdentity_Valid` (`internal/identity/identity_test.go:337`), `TestMagicLinkSeam_VerifyIdentity_NeverCommandLine` (`internal/auth/seam_test.go:148`). | Absorbed. |
| UNI-H4 (UNI:25) | CSRF is missing from the security requirements in the pass that births the first cookie. | BS:560-564; R15 (spec:494-525); spec:1212-1236. | As SEC-M2: `internal/serve/serve.go:173`, `internal/serve/origin.go:103`. Tests: `TestOriginCheckCoversSubtreeAndFallthrough` (`internal/serve/serve_test.go:983`), `TestNewServeShell_RosterPOSTWithoutOriginSignalIsDenied` (`cmd/dubplate-server/serve_test.go:805`). | Absorbed. |
| UNI-H5 (UNI:28) | Every deploy restarts the unit, so a memory-only nonce set forgives replay routinely. | BS:521-526; R11 (spec:430-452). | As SEC-M8: `internal/auth/nonce.go:155`. Tests: `TestInsertRefusesReplayAcrossRestart` (`internal/auth/nonce_test.go:105`), `TestConcurrentInsertAdmitsOne` (`internal/auth/nonce_test.go:542`). | Absorbed. |
| UNI-H6 (UNI:31) | Revocation by secret rotation plus a 90-day session is weak, and it forecloses per-device revocation. | BS:537-541; R4 (spec:337-347); R35 (spec:716-724). | `EpochStore`, `internal/auth/epoch.go:66`; `Read`, `internal/auth/epoch.go:175`; `Subscribe`, `internal/auth/epoch.go:207`. Tests: `TestCheckSession` (`internal/auth/epoch_test.go:74`), `TestMagicLinkSeam_Read` (`internal/auth/seam_test.go:274`), `TestSubscribeNotifiesOnBump` (`internal/auth/epoch_test.go:282`). | Absorbed, residual. `EpochStore.Bump` has no production caller until the revoke verb (STATUS:58-59; the 6d plan, `docs/plans/2026-09-26-rung-6d-identity-serve-store.md:93-109`). |

## UNI mediums (7)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| UNI-M1 (UNI:34) | The build order mixes the universal subsystem with the irreversible cutover and names no rollback. | By design: the pass split (spec:2508-2535, R23 at spec:579-585) keeps the cutover out of rung 5; rung 16 runs the ingress flip (spec:1866-1868); the 5d deploy carries a written rollback (plan:128-160) and checks by default (R-5d16, plan:593-606). | `deploy.sh` writes only under `--apply`, `scripts/deploy.sh:19`, parsed at `scripts/deploy.sh:111`. No test: shell scripts are ops tooling with no behavior tests (CLAUDE.md, "Stack rules"). | Absorbed, residual. A deliberate `--apply` by an agent is prevented only by its dispatch (plan:603-606). The cutover is rung 16's. |
| UNI-M2 (UNI:37) | Enumeration kindness is fixed behavior, argued from roster size. | R25 (spec:597-599); fork 4 (spec:248-251). | `unknownAddressAnswerThreshold`, `cmd/dubplate-server/auth.go:210`. Tests: `TestValidateMagicLinkSource_UnknownAddressAnswer` (`cmd/dubplate-server/auth_test.go:709`), `TestRequestLink_GenericAnswersAlike` (`internal/auth/handlers_test.go:600`), `TestRegisterAuthConfig_ThresholdDefaults` (`cmd/dubplate-server/auth_test.go:179`). | Absorbed. |
| UNI-M3 (UNI:40) | Two config fields would describe the streamer's address with different meanings. | BS:600-611 (the streamer's address is configured once, in driver config). | `DUBPLATE_PUBLIC_NAVIDROME_URL` is superseded, `internal/config/keys.go:137`, and `Load` refuses a superseded name, `internal/config/config.go:285`. Test: `TestLoad_ContractTable` (`internal/config/config_test.go:24`) proves the refusal on other superseded names; no test names this row. | Absorbed, residual. The driver's one address field is rung 12's (BS:600-611; spec:1784-1787). |
| UNI-M4 (UNI:43) | Access leaves the wall with no enable flag, which leaves two correlated keys silently optional. | By design: BS:587-591 (a front is an integration with an enable flag); spec:1645-1653 (a front's keys arrive with the rung that builds one). | Both Access keys are superseded, `internal/config/keys.go:133`; `rejectNamespace` refuses a superseded name, `internal/config/config.go:285`. Test: `TestLoad_ContractTable` (`internal/config/config_test.go:24`), over other superseded names. | Absorbed. |
| UNI-M5 (UNI:46) | The verify-identity seam is written in the present tense and does not exist. | BS:548-550; spec:1238-1250. | `Seam`, `internal/auth/seam.go:23`; `MagicLinkSeam` satisfies it, `internal/auth/seam.go:77`. Test: `TestMagicLinkSeam_VerifyIdentity` (`internal/auth/seam_test.go:73`). | Absorbed. |
| UNI-M6 (UNI:49) | Driver provisioning is stated as core behavior, and the first-user-becomes-admin hazard is parked in a runbook. | BS:592-599 (the destination is a driver capability with a core default); fork 5 (spec:253-262); spec:1268-1277. | Consumption answers with no destination, `internal/auth/handlers.go:356`; dubplate's own first admin is the `member` verb, as SEC-m9. | Carried to rung 12: the driver declares the destination and its provisioning, and the streamer-admin precondition travels there (spec:2849-2852; HISTORY:3416-3419). |
| UNI-M7 (UNI:52) | The bypass paths sit in core routing while the header name sits with the driver; they are one fact set. | spec:2849-2852 (the driver owns the header, injection, bypass paths, logout, and destination); spec:796-801. | `NewStrip` takes the driver's header as data, `internal/serve/strip.go:300`; the fallthrough slot is empty, `internal/serve/serve.go:140`. Tests: `TestStripWrapRemovesEveryIdentitySpelling` (`internal/serve/strip_test.go:85`) strips a driver-supplied name; `TestEmptyFallthroughAnswers404` (`internal/serve/serve_test.go:261`). | Absorbed, residual. The driver's declarations are rung 12's (spec:2849-2852). |

## UNI lows (3)

| ID | Finding | Absorbed by | Proof at HEAD | Status |
| --- | --- | --- | --- | --- |
| UNI-L1 (UNI:55) | The streamer's logout URL is unwired, so logout is half done. | spec:2849-2852 (the logout wiring lives with the driver). | dubplate's own sign-out clears the cookie, `internal/auth/handlers.go:381`. Test: `TestSignOut_ClearsCookie` (`internal/auth/handlers_test.go:954`). | Carried to rung 12 (spec:2849-2852). |
| UNI-L2 (UNI:58) | The repo already breaks the no-site-value rule, and the login page's masthead is the next breach. | spec:2843-2848; BS:406-412 (the estate's operations files are its operational record, not an adopter's kit). | The masthead reads the configured site name. Test: `TestNewWebRoutes_SiteNameMeta` (`cmd/dubplate-server/web_test.go:1681`). | Absorbed, residual. `config/navidrome.toml:13` still carries the estate's welcome string, an estate operations file under BS:406-412 that retires with Navidrome at 0.8 (CLAUDE.md, "Versions"). |
| UNI-L3 (UNI:61) | "One compact pass" understates the work and omits the plan-header discipline. | By design: R23 (spec:579-585) and the pass split (spec:2508-2535). | No code: the finding is about process. Every rung 5 plan header carries a ceiling and a checkpoint interval (plan:16-17). | Absorbed. |

## Findings absorbed only by R21

R-5d17 moves the forward-auth verify endpoint to rung 12 and requires that any finding absorbed only
by R21 be carried there (plan:624-625). Of the 50 findings, one rests on R21 alone: UNI-H2, the
forward-auth shape. Its row is carried to rung 12. SEC-M13 touches a forward-auth front too, and
the identity-source selector absorbs it, so its row stays absorbed with the front exemption's carry
named.

## The residue

This section lists every residual in one place, each with where it is recorded. Items 1 to 5 are
the plan's own named residuals. Items 6 to 9 are gaps this table found, recorded here for the
first time and not yet routed to a rung.

1. **The probe route under an unconfigured same-host proxy.** A same-host proxy that the operator
   did not configure as trusted forwards internet traffic from loopback, so that traffic reaches
   the probe route and learns the counts of its own headers the strip left. Recorded at plan:347
   and in Task 7's runbook outline (plan:1182-1183).
2. **Secrets, roster, and epoch integrity under the shared `music` uid.** Every `dubplate-server`
   verb reads `server.env`, so the shadow holds both HMAC keys and the SMTP password. The same uid
   owns `roster.json` and the epoch store, so any code that runs as `music` can rewrite them. The
   deploy's beets install (`uv tool install` from PyPI, run as `music` outside any sandbox) is one
   such path: it could grant itself the admin flag with no audit record, or delete an epoch record
   and undo a revocation (the rung 5d security review's L5). Carried to rung 16 (plan:140,
   plan:350).
3. **Group 2 survives from a trusted peer.** From a configured proxy, every forwarded-address
   header is kept, the configured one and every other (`internal/serve/strip.go:460`). Carried to rung
   12: strip every group 2 name but the configured header from a trusted peer (plan:1223-1224;
   fold record:153).
4. **The SMTP transport is not exercised on the box.** Carried to rung 16 (plan:1458).
5. **A deliberate `deploy.sh --apply` by an agent** is prevented only by its dispatch prompt
   (R-5d16, plan:603-606).
6. **Two SEC strip spellings have no decision.** SEC-M4 names `X-Remote-User` and
   `X-Authenticated-User`. Neither is in `identityExtras` (`internal/serve/strip.go:112`), and the
   list's admission record (`internal/serve/strip.go:26-111`) names neither as admitted or dropped.
   Recorded here; unrouted.
7. **The streamer-deprovision runbook step has no home.** The spec requires it (spec:1100-1105),
   and Task 7's outline (plan:1155-1201) omits it. It matters once SSO exists. Recorded here;
   unrouted.
8. **A header-strip hit is not logged as its own auth event.** BS:571-573 lists it among the auth
   events. The shell's only record of an identity header is the refusal log, which fires on the
   shell's own refusals (`internal/serve/serve.go:427`). Recorded here; unrouted.
9. **No WAF or Turnstile fallback is recorded.** SEC-m10 asks that one stay available for the login
   endpoint. No spec, plan, or ledger names it. Recorded here; unrouted.

The other residuals in the table, each recorded where its row cites:

- Navidrome's `/share/*` links stay live until 0.8 (BS:68-78).
- `FrontExemption`'s dual-stack peer match, to rung 12 (STATUS:56; arch reads:12-16).
- The estate's trusted-proxy mode stays direct until the cutover pins the tunnel container's
  address (plan:567-585).
- The magic link's token reaches a mail scanner and browser history (spec:1990-1994).
- Percent-encoding strictness, revisited at rung 12 (spec:2139-2142).
- No production caller of `EpochStore.Bump` until the revoke verb (STATUS:58-59).
- The per-address limiter's untracked admission past capacity (HISTORY:3412-3415).
- Rung 12's principal re-validation and its non-admin-to-streamer-admin refusal (HISTORY:2373-2375,
  HISTORY:3416-3419).
- The kind answer's membership and admin-contact disclosure (spec:1994; plan:750-752).
- The old-browser `Referrer-Policy` gap (HISTORY:2379).
- The proxied surface's headers and cache rule, the header-step reconciliation, and the
  shared-origin residual, to rung 12 (spec:2038-2046; HISTORY:2366-2367; HISTORY:2376-2377).
- `/rest/*` as an anonymous surface authenticated by the streamer alone (spec:1985-1987).
- The estate welcome string in `config/navidrome.toml:13` (BS:406-412).

## Verification

The row count and each review's own count, read from the texts:

```
grep -c '^\*\*\[blocker\]' docs/research/2026-09-02-login-spec-reviews/security.md    # 6
grep -c '^\*\*\[major\]'   docs/research/2026-09-02-login-spec-reviews/security.md    # 13
grep -c '^\*\*\[minor\]'   docs/research/2026-09-02-login-spec-reviews/security.md    # 12
grep -c '^\*\*\[Critical\]' docs/research/2026-09-02-login-spec-reviews/universality.md # 3
grep -c '^\*\*\[High\]'     docs/research/2026-09-02-login-spec-reviews/universality.md # 6
grep -c '^\*\*\[Medium\]'   docs/research/2026-09-02-login-spec-reviews/universality.md # 7
grep -c '^\*\*\[Low\]'      docs/research/2026-09-02-login-spec-reviews/universality.md # 3
grep -cE '^\| (SEC|UNI)-' docs/research/2026-09-25-rung-5-sec-uni-disposition.md      # 50
```

Every cited test was listed with `go test -list` against its package at `03e1518`, and each name
resolved. The task report carries the output.
