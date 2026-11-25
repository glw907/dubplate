// Package ops declares the operation vocabulary a caller reaches the
// server's work through, and nothing that runs one: the admin and
// submission interfaces, the background-job vocabulary, their wire
// types, the error vocabulary, and the privilege each operation
// requires.
//
// [Admin] holds the member and settings operations and [Submissions]
// the submission operations. Both take the acting [identity.Actor] as a
// parameter of every method, never as a field of a request type, so a
// caller can never widen its own privilege by shaping a request body.
//
// Every request and response type is JSON-serializable with explicit
// snake-case tags, the source of the API's wire shape once a driver
// exposes it over HTTP. A failure is an [*Error] or wraps one: a caller
// branches on its [Code] with [errors.Is] or [CodeOf], and a driver
// answers with its Reason, Detail, and Rejects.
package ops
