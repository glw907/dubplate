// Package ops declares the operation vocabulary a caller reaches a
// long-running job through, and nothing that runs one: the interface,
// its wire types, the error vocabulary, and the catalog that reports
// which operation is available and to whom.
//
// [Seam] takes the acting identity as a parameter of every method,
// never as a field of a request type, so a caller can never widen its
// own privilege by shaping a request body. An [Actor] built here
// names a local operator only; a network caller's kind lands with the
// projection that first verifies one.
//
// Every request and response type is JSON-serializable with explicit
// snake-case tags, the source of the API's wire shape once a driver
// exposes it over HTTP. A caller matches a failure with [errors.Is]
// against one of this package's sentinels, never by inspecting an
// [*Error]'s fields directly.
package ops
