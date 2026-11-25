// Package ops declares the operation vocabulary a caller reaches the
// server's work through, and nothing that runs one: the admin and
// submission interfaces, the background-job vocabulary, their wire
// types, the error vocabulary, and the privilege each operation
// requires.
//
// [Admin] holds the seven member and settings operations and
// [Submissions] the five submission operations. Both take the acting
// identity as a parameter of every method, never as a field of a request
// type, so a caller can never widen its own privilege by shaping a
// request body. An [identity.Actor] is local, built from an
// operating-system username, or network, built from a verified
// [identity.Identity] and its session's anchor; the network kind
// renders its durable UserID, never its principal.
//
// Every request and response type is JSON-serializable with explicit
// snake-case tags, the source of the API's wire shape once a driver
// exposes it over HTTP. A failure is an [*Error] or wraps one: a caller
// branches on its [Code] with [errors.Is] or [CodeOf], and a driver
// answers with its Reason, Detail, and Rejects.
package ops
