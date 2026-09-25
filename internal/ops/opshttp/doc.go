// Package opshttp projects the ops seam onto HTTP: every operation is
// a POST to its catalog name, carrying the operation's request type as
// one JSON object and answering with its response type or a JSON error
// naming one of the seam's wire codes.
//
// Authorization happens in one place, a decorator over the seam that
// reads each operation's privilege from package ops, asks the
// authorizer it was built with, and calls through only with the actor
// that authorizer answered. The HTTP layer never reads a header,
// cookie, or query parameter for identity, and it passes the zero
// actor into the decorator, which discards whatever actor its caller
// passes. A verified identity reaches the authorizer only through the
// request's context.
//
// Nothing here is exported: the package has no caller until a server
// mounts it behind the session middleware that puts a verified
// identity in that context.
package opshttp
