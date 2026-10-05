# Rung 5d-i close: architecture reads

Two `go-architecture-reader` dispatches (Opus, no plan in context) at `7ca2f57`, 2026-10-05.
Both graded their package "workmanlike" and found no correctness gap on any production path.
Per the 5d plan, findings route to the architecture-cleanup pass; the one latent bug routes to
rung 12, the hook's first consumer.

## internal/serve

Density 0.87 comment-to-code (1312/1514) against `net/http/httputil` at 0.45.

Latent bug, routed to rung 12. `FrontExemption` and `NewFrontExemption` (`strip.go:232-261`, exempt
branches at `:360-372`, `:449`, `:502-531`) have no production caller: `cmd/dubplate-server/serve.go`
passes `serve.FrontExemption{}`. Its `proxies map[string]struct{}` is matched against the raw
`PeerAddress` string, while `TrustedProxyConfig.trusts` (`clientaddr.go:188`) parses and unmaps. On
a dual-stack listener an IPv4 front arrives as `::ffff:a.b.c.d` and would never match. The reader
recommends deleting the type and its branches now and letting rung 12 rebuild it keyed on unmapped
`netip.Addr` through `TrustedProxyConfig.trusts`.

Cleanup findings:

- Exports ahead of a caller: `SetFallthrough` (`serve.go:140`) and the fallthrough fields; `NewStrip`'s
  `header` parameter (`strip.go:300`), always `""` in production.
- Unexport (in-package or test-only references): `Listen`, `PeerAddress`, `RewriteDocument`,
  `Normalize`, `Prefix.Owns`, `Prefix.Route`, `Strip.Wrap`, `ModeTrustedProxy` and `ProxyMode`,
  `ThemeCookieName`. `ErrInvalidPrefix` and `ErrInvalidBindAddr` are reached only by cmd tests.
- Duplication: `prefix.go:85-95` restates `normalize.go:119-129`'s segment loop (extract
  `segmentFault`); `clientaddr.go:255-285` hand-rolls `strings.Split` and `strings.Trim`;
  `strip.go:365` rebuilds `identitySpellings` per request and `refusedProxyHeaders`
  (`clientaddr.go:68`) rebuilds its map per call (precompute both).
- File split: two disagreeing package doc comments (`doc.go:1`, `clientaddr.go:1-3`); `document.go`
  mixes four concerns (split into document, assets, themecss, log files).
- Test seam: `newHTTPServer`'s `writeTimeout` parameter (`server.go:77-92`) exists for one test.
- Density: 93 process citations; the header bibliography at `strip.go:26-111` belongs in a research
  doc.

Best passages: the trailer scrub and `Close` override (`strip.go:352-358`, `:432-438`); the one-rule
tightening over `http.CrossOriginProtection` (`origin.go:50-55`, `:103-116`); `statusCapture`
forwarding `io.ReaderFrom` and shutdown on the failure path (`serve.go:375-386`, `server.go:177-183`).

## cmd/dubplate-server

Density 0.419 (1487/3550) against `cmd/go/internal/work` at 0.260.

Cleanup findings, ranked by the reader:

- Key-name duplication (the largest cost): the magic-link source spells eleven keys as literals
  (`auth.go:46-60`, `:268-273`) while `serve.go:25-37` declares them as constants; `doctor.go:199`
  redeclares the roster key; `DUBPLATE_BACKUP_REMOTE` appears four times, `DUBPLATE_STATE_DIR`
  three. One const per key, as `proxy.go:14-17` does. A rename can drift silently today.
- The trusted-proxy list is parsed three times (`proxy.go:58`, `doctorprobe.go:63`, `:287`):
  `resolveTrustedProxy` should return the parsed addresses once.
- `doctor.go:437-466` reimplements config's `resolveIdentitySource` and `wall.get`; export one
  `config.LookupEnv` and `(IdentitySource).Values` instead (an `internal/config` change, so it
  waits for a pass that may touch config).
- Two owner checks and uid resolvers (`member.go:196-223`, `unidecode.go:271-310`); two restated
  unstamped-commit refusals (`shadow.go:267`, `:373`); `methodOnly` (`serve.go:397`) duplicated by
  `newStaticFiles`'s 405 arm (`web.go:431-435`).
- Test-only seams: `newServeShell`'s returned sweep channel (`serve.go:254`, `:330`);
  `newWebRoutesFromFS` (`web.go:226-231`); `authHandlersConfig` (`serve.go:372`).
- `ShadowIntegration` (`shadow.go:48`) exported in package main.
- File split: `web.go` (848 lines) holds library machinery that belongs in `internal/serve` or its own
  package; doctor rows live in `store.go:209-442` and `backup.go:383-572`.
- About 30 repeated `printRow` error checks (`doctor.go:116-181`, `shadow.go:449-513`): a sticky-error
  writer or a row slice; `registerShadowIntegration` and `registerShadowThresholds` merge.
- Density: 84 process-citation lines; `serve.go:203-253` is a 50-line doc on an unexported function.

Best passages: the probe's pointer-decoded counts (`doctorprobe.go:371-379`); `readOperatorCSS`'s
checked-is-read file handling (`web.go:658-709`); `awaitBackup`'s nil-channel loop and
`WithoutCancel` cancellation (`backup.go:296-312`, `:333-340`).
