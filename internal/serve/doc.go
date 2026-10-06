// Package serve implements the dubplate HTTP shell: the root
// composition every request passes through before reaching a
// registered handler.
//
// # Composition order
//
// [Shell.Handler] composes six steps, in this order, each wrapping the
// next. The order is a property of the chain, never a feature of the
// mux: no step is registered as a route, and no test in this package
// proves the order by reading the chain's construction.
//
//  0. The security headers (securityHeaders). X-Content-Type-Options:
//     nosniff, Referrer-Policy: no-referrer, and, once
//     [Shell.EnableHSTS] is called, Strict-Transport-Security. The
//     step wraps everything, so a refusal by any later step, a 404
//     included, carries them. A document's Content-Security-Policy is
//     not among them: it is per document, set by the handler that
//     serves it ([DocumentCSP]).
//  1. The ingress strip ([Strip.Wrap]). It wraps everything, so no
//     route is reachable with an identity header, a forwarded-address
//     header, or a Connection-listed header still attached.
//  2. Path normalization ([Normalize]). It refuses six shapes, each
//     with a status and no partial routing, before any later step runs
//     at all.
//  3. The echo method refusal (refuseEchoMethods). TRACE and TRACK are
//     refused here with a 405, before the origin check can claim them
//     with a 403.
//  4. The origin check (originCheck). It judges only a path
//     normalization has admitted, and covers every registered subtree
//     and every path outside the prefix alike, so no mount can forget
//     to install it.
//  5. Subtree dispatch. A path a registered subtree claims reaches
//     that subtree; the claim is segment-aware, on case-sensitive
//     bytes, never a plain string trim. Every other path, outside the
//     owned prefix or under it, answers 404 from
//     [http.NotFoundHandler].
//
// # The extension point
//
// A consumer registers a subtree under the owned prefix by name, at
// construction time, through [Shell.Register]; the shell owns the mux
// the registration lands in, and two registrations of the same name
// panic, naming the subtree. The method takes an [http.Handler], the
// stdlib's, and no interface of this package's own.
package serve
