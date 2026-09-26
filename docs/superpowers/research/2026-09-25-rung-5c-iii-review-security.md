# 5c-iii security review (web-auth-security-reviewer)

Range `6e6fd12..HEAD` over `internal/serve` and `cmd/dubplate-server`, plus `web/scripts/rendered/`.
Nothing Critical or High.

- **M1 (medium).** `web/scripts/gates/root-absolute.ts:27-33` catches only literal URLs in `web/src`;
  sinks such as `location.href`, `location.assign`/`replace`, `window.open`, `new URL('/x', location)`,
  `new Request`, `sendBeacon`, `formaction=`, `srcset=`, URLs held in variables, and built JS from
  dependencies escape it. At rung 12 the fallthrough is Navidrome and a stray POST body (a token, an
  address) would reach its logs; CSP `'self'` admits it. Fix: assert "no request outside the prefix"
  on every rendered row from the existing request log (`run.ts:820`, `:894` do it only for icon rows),
  and extend the source patterns and a built-JS scan. Matters most before Tasks 7-8.
- **M2 (rung 12).** `internal/serve/csp.go:33-37,73`: `script-src 'self'`, `connect-src`, and
  `form-action` admit the same-origin fallthrough streamer. At rung 12, scope them to
  `<origin><prefix>/` (and `/_app/` for scripts) from `DUBPLATE_PUBLIC_URL`; keep the prefix
  redirect-free.
- **L1.** `internal/serve/document.go:148-152` inserts operator text (site name, prefix) before
  counting the placeholder; a site name containing `__DUBPLATE_THEME__` breaks or masks the
  one-placeholder check. Fix: `RewriteDocument` first, then insert; refuse the token in `SiteName`
  and the prefix.
- **L2.** `DUBPLATE_SITE_NAME` has no grammar (`internal/config/config.go:352`, `keys.go:23`). Add
  valid UTF-8, no C0/DEL/C1, no bidi overrides, at most about 100 bytes, no placeholder. Task 7 must
  render it by text interpolation, never `{@html}`.
- **L3.** `cmd/dubplate-server/web.go:566` trusts a `#` prefix for the branding neutral, and the
  neutral is parsed twice (`web.go:530-537`). One parse; check against the exported `hexColorRE`.
- **L4.** SVG responses (`web.go:430-433`) carry no CSP. Send `default-src 'none';
  style-src 'unsafe-inline'; sandbox` on SVGs from `newStaticFiles`.
- **L5.** `cmd/dubplate-server/rendered_test.go:428` passes the whole environment (secrets included)
  to Node and Playwright on the workstation. Pass an allowlist.
- **I1-I4.** `checkVariantCSP` discards each variant's policy rather than asserting it equals the
  served header (`web.go:371`); `headAnchor` can anchor outside `<head>` (`document.go:163-174`),
  so require `anchor < index("</head>")`; operator CSS is unlayered by design; the harness request
  log serves captured headers on loopback under the build tag only.

Checked clean: head escaping (`html.EscapeString`, double-quoted), theme-color validation, CSP
hashing and recheck, the branding stylesheet grammar, the fonts and icons routes (exact map after
Normalize, fixed Content-Type, `nosniff`, strong ETag, GET/HEAD only), the runtime base, the token
URL's referrer handling, and the theme cookie.
