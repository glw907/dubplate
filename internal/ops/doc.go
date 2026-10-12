// Package ops declares the operation vocabulary a caller reaches the
// server's work through, and nothing that runs one: the two seams, their
// wire types, the error vocabulary, and the catalog that reports which
// operation is available and to whom.
//
// [Seam] holds the four job operations and [AdminSeam] the seven member
// and settings operations. Both take the acting identity as a parameter
// of every method, never as a field of a request type, so a caller can
// never widen its own privilege by shaping a request body. An [Actor]
// is local, built from an operating-system username, or network, built
// from a verified [identity.Identity] and its session's anchor; the
// network kind renders its durable UserID, never its principal.
//
// Every request and response type is JSON-serializable with explicit
// snake-case tags, the source of the API's wire shape once a driver
// exposes it over HTTP. A caller matches a failure with [errors.Is]
// against one of this package's sentinels, never by inspecting an
// [*Error]'s fields directly, except that a client branches on
// [Error.Detail] for a refusal finer than its code.
package ops
