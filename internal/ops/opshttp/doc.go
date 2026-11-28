// Package opshttp projects [ops.Admin] and [ops.Submissions] onto HTTP:
// every operation is a POST to its name, carrying the operation's
// request type as one JSON object and answering with its response
// type or a JSON error naming one of package ops's wire codes, with an
// optional detail finer than the code.
//
// [NewMount] serves the operations under <prefix>/api/v1/ for the
// sessions a [Verifier] admits. Authorization happens in one place,
// before a route reads a body: it reads the operation's privilege from
// package ops, admits the call only when the verified caller's granted
// privilege reaches it, and calls the operation only with the actor the
// verifier answered. The HTTP layer never reads a header, cookie, or
// query parameter for identity; the verifier alone does.
package opshttp
