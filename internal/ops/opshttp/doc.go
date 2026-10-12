// Package opshttp projects the ops job seam and admin seam onto HTTP:
// every operation is a POST to its catalog name, carrying the
// operation's request type as one JSON object and answering with its
// response type or a JSON error naming one of the seams' wire codes,
// with an optional detail finer than the code.
//
// Authorization happens in one place, a decorator over both seams that
// reads each operation's privilege from package ops, asks the
// authorizer it was built with for the caller's actor and granted
// privilege, admits the call only when that grant reaches the
// operation's privilege, and calls through only with the actor the
// authorizer answered. The HTTP layer never reads a header,
// cookie, or query parameter for identity, and it passes the zero
// actor into the decorator, which discards whatever actor its caller
// passes. A verified identity reaches the authorizer only through the
// request's context, where the mount puts the caller its verifier
// admitted.
//
// Nothing here is exported: the package has no caller until a server
// mounts it behind the session middleware that puts a verified
// identity in that context.
package opshttp
