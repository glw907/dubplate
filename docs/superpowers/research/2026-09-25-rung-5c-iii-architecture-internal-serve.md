# `internal/serve` architecture read (5c-iii merge side)

Read at HEAD `db1e111`, no plan in context, per CLAUDE.md's "Reviewers" (one
`go-architecture-reader` per touched Go package at each pass's merge).

**Verdict: workmanlike.** Density 0.893 against `net/http`'s 0.462.

## Findings

- **`NewTrustedProxyConfig` has no production caller.** `internal/serve/clientaddr.go:25`
  constructs a `TrustedProxyConfig`, but `cmd/dubplate-server/serve.go:244` declares a zero
  value, so trusted-proxy mode cannot turn on in the shipped binary. `ProxyMode`, `ModeDirect`,
  `ModeTrustedProxy`, and `String` (`clientaddr.go:36-52`) are unused as a result. **Routed:**
  5d-i Task 1, the trusted-proxy class; its Task 0 must read this finding before wiring it.
- **Rung 12 hooks shipped early.** `SetFallthrough` (`serve.go:140`, fields `:31-32`),
  `NewFrontExemption` (`strip.go:242`, no caller), and `NewStrip`'s header parameter
  (`strip.go:300`, always `""`) exist ahead of rung 12's fallthrough streamer.
  `NewStrip`'s `sessionCookie` is unread by `Strip`; only `Shell` reaches it
  (`s.strip.sessionCookie` at `serve.go:242`) — move it to `NewShell`. **Routed:** rung 12, or
  unexported now by the architecture cleanup pass.
- **Candidates to unexport.** `RewriteDocument` (`document.go:31`), `ThemeCookieName`
  (`document.go:21`), `Listen` (`server.go:113`), `Normalize` (`normalize.go:83`),
  `Prefix.Owns`/`Route` (`prefix.go:116,149`), `Strip.Wrap` (`strip.go:333`), `PeerAddress`
  (`clientaddr.go:68`). Sentinels `ErrInvalidPrefix`/`ErrInvalidBindAddr` are only test-branched,
  which is defensible for an error sentinel.
- **Test-only seams.** `addBearerBypass` plus the bypass mux (`origin.go:37-91`);
  `newHTTPServer`'s `writeTimeout` parameter (`server.go:92`); the `deny` field
  (`origin.go:58`).
- **Duplication.** The segment walk in `prefix.go:85-95` restates `normalize.go:119-129`;
  peer-set conversion and membership checks appear in three places across `clientaddr.go` and
  `strip.go`; `splitHops`/`trimHop` are hand-rolled (`clientaddr.go:139-170`);
  `methodGetHead` is restated in `cmd/dubplate-server/web.go:431-436`; static serving exists in
  two packages; per-attribute regexes repeat at `csp.go:40-42` and `document.go:348-350`.
- **File split.** `document.go` holds five concerns in one file: rewrite, the handler, static
  serving, logging, and CSS scanning. Two files carry their own package comment (`doc.go` and
  `clientaddr.go`).
- **Comments.** About 107 process-citation lines (a rung, task, or ruling id in a shipped
  comment — the staged gate leg this repo is building toward). Research-history narration sits
  at `strip.go:26-111`, `serve.go:295-303`, `strip.go:193-198`, `prefix.go:77-84,107-115`, and
  `server.go:77-83`.

## Best code in the package

- `document.go:149-163` — placeholder ordering in the document rewrite.
- `strip.go:412-443` — trailer scrubbing.
- `origin.go:103-116` — the `CrossOriginProtection` tightening.

## Routing

Trusted-proxy wiring to 5d-i Task 1. Rung 12 hooks recorded for rung 12, or unexported now by
the cleanup pass. Everything else to the architecture cleanup pass.
