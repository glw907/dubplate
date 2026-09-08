// Package serve implements the dubplate HTTP shell (PA:79): the root
// composition every request passes through before reaching a
// registered handler or the fallthrough.
//
// # Composition order
//
// [Shell.Handler] composes four steps, in this order, each wrapping the
// next. The order is a property of the chain, never a feature of the
// mux: no step is registered as a route, and no test in this package
// proves the order by reading the chain's construction.
//
//  1. The ingress strip ([Strip.Wrap]). It wraps everything, so no
//     route is reachable with an identity header, a forwarded-address
//     header, or a Connection-listed header still attached.
//  2. Path normalization ([Normalize]). It refuses six shapes, each
//     with a status and no partial routing, before prefix routing runs
//     at all.
//  3. Prefix routing ([Prefix.Route]). The one owned prefix is matched
//     segment-aware, on case-sensitive bytes, never by a plain string
//     trim.
//  4. The fallthrough. Empty at rung 5, which is what makes rung 5 safe
//     to ship without a driver. Rung 12's reverse proxy occupies this
//     point, and no other: by the one-owned-prefix rule the streamer
//     owns everything outside it, so the driver is the root mux's
//     fallthrough rather than a subtree mounted under the prefix.
//
// # The extension point
//
// A consumer registers a subtree under the owned prefix by name, at
// construction time, through [Shell.Register]; the shell owns the mux
// the registration lands in, and two registrations of the same name
// panic, naming the subtree. A consumer sets the one root-level
// fallthrough slot through [Shell.SetFallthrough]; a second call panics.
// Neither method takes or returns an interface of this package's own:
// both take [http.Handler], the stdlib's, which is the whole point of
// an extension point built for a rung 5 with no driver yet to bring one.
//
// The shell's own session cookie (R-E) is removed from any request
// reaching the fallthrough, so rung 12's proxy never forwards it to the
// streamer; a shell built with no cookie name removes nothing.
package serve
