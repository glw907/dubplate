// Package serve implements the dubplate HTTP shell: the root
// composition every request passes through before reaching a
// registered handler.
//
// # Composition order
//
// [Shell.Handler] composes these layers, outermost first, each wrapping
// the next. No layer is registered as a route, so every path the
// server answers passes through all of them.
//
//  1. The security headers: X-Content-Type-Options: nosniff,
//     Referrer-Policy: no-referrer, and, once [Shell.EnableHSTS] is
//     called, Strict-Transport-Security. A refusal by any later layer,
//     a 404 included, carries them. A document's
//     Content-Security-Policy is set by the handler that serves it.
//  2. The refusal record: one log record for each request a later
//     layer refused.
//  3. The direct-mode warning, logged once, the first time a loopback
//     or private peer sends a forwarded-address header.
//  4. The ingress strip, which removes every identity header, every
//     forwarded-address header from an untrusted peer, and every header
//     the request's Connection header lists.
//  5. Path normalization, which refuses an unsafe path shape.
//  6. The echo-method refusal: TRACE and TRACK answer 405, before the
//     origin check can answer them 403.
//  7. The origin check, which refuses a cross-site state-changing
//     request.
//  8. Subtree dispatch. A path a registered subtree claims, on a
//     segment boundary, reaches that subtree; every other path answers
//     404.
//
// A consumer registers each subtree by name through [Shell.Register]
// before calling [Shell.Handler].
package serve
