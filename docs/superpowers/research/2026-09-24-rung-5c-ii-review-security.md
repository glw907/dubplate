# Rung 5c-ii pass-end review: web-auth security

Reviewer: `web-auth-security-reviewer` (Opus 5.5), over `0c2745f..2424b9a` in `internal/auth`,
`internal/identity`, `internal/serve`, and `cmd/dubplate-server/{serve,web}.go`. No critical or
high finding; two medium, six low, five info. Two points confirmed with a scratch Go program.

## Medium

- **M1. `readOperatorCSS` follows symlinks and checks twice** (`cmd/dubplate-server/web.go:323-351`).
  `os.Stat` follows the link, so an `x.css` symlink to `roster.json` reports a regular file; the
  suffix rule tests the configured string, not the target; the roster passes every content
  check; `<prefix>/operator.css` would serve every member address and admin flag unauthenticated.
  Stat and Open are separate calls, so a FIFO swapped in between hangs startup and another
  regular file bypasses the check. Fix: open once with
  `O_RDONLY|O_NOFOLLOW|O_NONBLOCK|O_CLOEXEC`, `f.Stat()` the descriptor and require a regular file;
  `filepath.EvalSymlinks` so parent-directory links are covered and require the resolved path to
  end in `.css`; refuse a file `os.SameFile` matches to the roster and anything under `StateDir`;
  optionally refuse a group- or world-writable file. Add symlink rows to R-5cii13's states and to
  `TestNewWebRoutesFromFS_OperatorCSSRefusals`. **Fix before any deploy.**
- **M2. The privileged action shares an origin with the fallthrough** (`cmd/dubplate-server/serve.go:301,308`).
  Sec-Fetch-Site, `CrossOriginProtection`, SameSite, HttpOnly, and the CSP all assume same-origin
  means dubplate's code. Once Navidrome is mounted at the fallthrough, any script on that origin
  (a stored XSS through member tag metadata, a share page) can read the roster and POST members
  with an admin's session. Not live today (`SetFallthrough` is not called). Carried to rung 12:
  separate origins (a subdomain) or step-up confirmation for roster writes, until 2.0 retires
  Navidrome.

## Low

- **L1.** `LogRequests` logs `r.Method` unbounded (`internal/serve/document.go:171-172`); a
  non-browser client can write about 1 MiB per line. Log a known method or `"other"`.
- **L2.** `isLoopbackHost` misses `LOCALHOST`, `localhost.`, and `*.localhost`, so HSTS is sent
  for them (`cmd/dubplate-server/auth.go:474-482`); lowercase, strip one trailing dot, match
  `localhost` and the `.localhost` suffix (RFC 6761), with `sendsHSTS` rows.
- **L3.** The CSP is computed over the raw document, but variants are rewritten
  (`web.go:194` against `:240`); a placeholder or anchor inside a script body would change a hashed
  script and fail closed. Recompute per variant or assert offsets fall outside script spans.
- **L4.** The gated members document is `private, no-cache` (`web.go:77`); use `no-store` to
  match the API.
- **L5.** The roster POST has no throttle (`internal/auth/admin.go:227`); a per-actor bucket.
- **L6.** n5 can refuse a pre-existing roster whole (`internal/identity/identity.go:75-78`); a
  doctor row naming the member's position and an upgrade note.

## Info

I1 path characters `;`, `=`, `@` pass `Normalize` (no minted URL uses them); I2 three 5c-iii
requirements for the login/link page (consume only on a click, `history.replaceState` to drop
`?t=`, no third-party load); I3 no `Permissions-Policy` or COOP, `style-src 'unsafe-inline'`
required by `app.html`; I4 `/roster` measurably slower than `/roster/x` (the route is public
anyway); I5 the branding error quotes an unknown name with `%q`, so a pasted secret would appear
in the startup error.

## Checked and sound

The admin flag is never taken from the request; refusals equal `http.NotFound` and set no cookie;
cross-site and header-less POSTs are refused; form parsing is capped before parse, PostForm-only,
with a closed field set; the branding grammar is anchored and escapes nothing into place; the
theme cookie only selects a precomputed variant; exact-path routing defeats cache deception; the
header step is outermost; `LogRequests` never sees the token (`TestNewServeShell_MintedLinkNeverLogsItsToken`).
