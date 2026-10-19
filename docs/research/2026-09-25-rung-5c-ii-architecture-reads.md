# 5c-ii architecture reads (go-architecture-reader, master 1a7cc9e, 2026-09-25)

Routing (conductor): every finding routes to the architecture cleanup pass except where noted. The forward hooks built ahead of a caller (serve's fallthrough, front exemption, trusted-proxy mode, bearer bypass; identity's Principal, DisplayName, Source; auth's epoch Bump and Subscribe) are each named for the rung that consumes them (12, 5d, the revoke verb); the cleanup pass removes them and that rung re-adds each with its caller, per the 2026-09-12 refinement of Build for 2.0. auth's revocation epoch has no production writer: revocation is checked but never recorded until the revoke verb lands.

## internal/auth — workmanlike (density 0.663 vs net/http/cookiejar 0.352)
- epoch.go:103,207,36 EpochStore.Bump/Subscribe/ErrEpochBackward: no production caller, so a deployment never writes an epoch and CheckSession/CheckLink always find none. Revocation has no write path yet (the revoke verb is its consumer). Delete Subscribe and its machinery; unexport Bump until then.
- ~40 exports with no outside caller (session, token, store, epoch, nonce, limit, roster sentinels and functions); the real API is ~22 identifiers (Open*, CreateRoster, NewLimiter, NewMagicLinkSeam, NewHandlers, NewRosterAPI, RequireAdmin, their config/value types, SessionCookieName, SweepInterval, RunSweep, ErrRosterMissing/Exists). Unexport the rest.
- seam.go:40,144 Seam.Read has no caller: one-method interface at the consumer.
- Test seams: RunSweep's sweep/now params (nonce.go:123); rosterAPI.writeWait field (admin.go:167-170); keyedLimiter.size (limit.go:265); OpenNonceStore's discarded startup NonceSweep (nonce.go:93).
- store_test.go:34-53 proves atomicfile's perm, not a store's.
- Duplication: RecordID+EpochPath label/dir agree by hand (store.go:83-146); seam.go:119-122 restates Member.principal; CheckSession/CheckLink reimplement Read (epoch.go:225-269); two rune predicates (roster.go:654-681); normalizeMemberAddress pass-through and repeated normalization; writeAnswer wrapper.
- File split: signed-value codec to signed.go; roster.go (867 lines) member rules to member.go.
- Comments: 54 process-citation lines, several stale (roster.go:56, session.go:85, epoch.go:217, roster.go:191); record-version rationale x4; dead-guard rebuttals (token.go:125-130, 296-301).
Best: lockRosterDir locks the directory; limiter eviction policies; generic answer responds before mail.

## internal/serve — workmanlike (density 0.90 vs net/http 0.46)
- Forward hooks with no production caller (rung 12 / 5d): Shell.SetFallthrough (serve.go:140) and the cookie removal it feeds (serve.go:240-326); NewFrontExemption and NewStrip's header parameter (strip.go:232-300); NewTrustedProxyConfig/ProxyMode (clientaddr.go:25-60); addBearerBypass (origin.go:50-115). Rung 12 and 5d decide; the cleanup pass may remove and let the consuming rung re-add.
- ~15 exports with no outside caller: Normalize, Strip.Wrap, Prefix.Route/Owns, Listen, RewriteDocument, ThemeCookieName, PeerAddress, ErrInvalidPrefix, ErrInvalidBindAddr: unexport.
- server.go:77-92 newHTTPServer's writeTimeout parameter is a test seam; set the field in the test.
- strip.go:272,300 Strip carries sessionCookie it never uses (courier for Shell).
- Duplication: refuse() x3 (serve.go:204, normalize.go:86, origin.go:76); proxy-trust test x3; segment walk (normalize.go:119-129 vs prefix.go:85-95); six tag/attr regexps (csp.go:40-42, document.go:256-259); exempt filtering twice (strip.go:366-453); splitHops/trimHop reimplement strings.Split/Trim; Terminator is http.NotFoundHandler().
- File split: document.go holds five concerns (static.go, markup/log files); serve.go four (log.go; cookie removal to strip.go); clientaddr.go:1-3 second stale package comment.
- Comments: 108 process-citation lines; stale: prefix.go:139 ("third step", is fifth), strip.go:268 ("outermost"), serve.go:44, doc.go:7/serve.go:148 ("six steps", is seven); strip.go:26-111 provenance log.
Best: logRefusal's context-slot keying; trailerScrubbingBody; branding's refuse-whole values.

## internal/identity — workmanlike (density 1.03 vs net/mail 0.21)
- identity.go:194,196,199 Identity.Principal/DisplayName/Source: one writer (auth/seam.go:127-129), no reader outside the package; Valid's Principal/Source terms cannot fail in production. Forward-hook fields ahead of a caller (the refined 2026-09-12 rule); reduce to Address+Channel, each field added with its first reader.
- identity.go:23 MaxAddressLength unreferenced outside its file: unexport.
- identity.go:29 ErrInvalidAddress matched only by auth tests: unexport or give it a production branch.
- identity.go:97-116 isDefaultIgnorableCodePoint derivation is subsumed by the Cf refusal (0 differing code points, go1.27.1 sweep); replace with unicode.In(..., Other_Default_Ignorable_Code_Point, Variation_Selector); delete TestDefaultIgnorableDerivedSetSize, which also breaks on each Unicode bump.
- identity.go:58-59,124-131 double scan (ContainsFunc then invisibleRune); one IndexFunc.
- identity_test.go:434-475 two tests that assert struct assignment; delete.
- comments: package doc process citations; rung-6 citations; refused-rune list stated three times; test comments carry process history.
Best: identity.go:61 re-parse fixed point; Channel iota+1 with Valid refusing ChannelCommandLine; the "é" length-cap fixture.
