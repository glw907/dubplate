// Package opshttp projects the admin and submission seams onto HTTP:
// every operation is a POST to its name, carrying the operation's
// request type as one JSON object and answering with its response
// type or a JSON error naming one of the seams' wire codes, with an
// optional detail finer than the code.
//
// Authorization happens in one place, the admission every route makes
// before it reads a body. It reads the operation's privilege from
// package ops, asks the authorizer it was built with for the caller's
// actor and granted privilege, admits the call only when that grant
// reaches the operation's privilege, and calls the seam only with the
// actor the authorizer answered. The HTTP layer never reads a header,
// cookie, or query parameter for identity. A verified identity reaches
// the authorizer only through the request's context, where the mount
// puts the caller its verifier admitted.
//
// Nothing here is exported: the package has no caller until a server
// mounts it behind the session middleware that puts a verified
// identity in that context.
package opshttp
