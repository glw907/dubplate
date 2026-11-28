package opshttp

import (
	"context"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/serve"
)

// A Verifier admits one request to the mount. It reads whatever
// credential the request carries, may set a refreshed credential on w,
// and answers the actor the call proceeds as and the privilege that
// actor holds. A refused credential wraps [ops.CodeUnauthenticated],
// which the mount answers 401 unauthenticated; any other verifier error
// is the server failing and answers 500. A verifier logs its own
// failures: the mount never writes an error's text, which may name a
// member.
type Verifier func(w http.ResponseWriter, r *http.Request) (identity.Actor, ops.Privilege, error)

// A route serves one operation to a caller the mount's verifier
// admitted.
type route func(w http.ResponseWriter, r *http.Request, actor identity.Actor, granted ops.Privilege)

// A mount serves each operation as POST <prefix>/api/v1/<name>. It
// dispatches by exact path itself rather than through [http.ServeMux],
// whose method-mismatch and not-found answers are plain text, so that
// every answer carries the JSON error shape.
type mount struct {
	base   string
	verify Verifier
	log    *slog.Logger
	routes map[ops.Op]route
}

// NewMount returns the handler for the cookie mount's subtree api,
// which the serve shell hands the full path, prefix included. It cuts
// <prefix>/api/v1/ itself rather than through [http.StripPrefix], whose
// mismatch answer is plain text, and answers every other path under api
// with a JSON not_found before it verifies anyone: the document is
// public, so the order reveals nothing, and the owned prefix ends here
// whatever a later route does.
//
// Every operation runs as the actor verify admits and never as a field
// of the request. prefix is the shell's owned prefix. NewMount fails on
// a nil admin or submissions implementation, verifier, or logger, and
// on the zero Prefix.
func NewMount(admin ops.Admin, subs ops.Submissions, prefix serve.Prefix, verify Verifier, log *slog.Logger) (http.Handler, error) {
	switch {
	case verify == nil:
		return nil, errors.New("verifier is nil")
	case prefix == serve.Prefix{}:
		return nil, errors.New("prefix is the zero Prefix")
	case log == nil:
		return nil, errors.New("logger is nil")
	case admin == nil:
		return nil, errors.New("admin operations are nil")
	case subs == nil:
		return nil, errors.New("submission operations are nil")
	}
	m := &mount{base: prefix.String() + "/api/v1/", verify: verify, log: log, routes: map[ops.Op]route{}}
	handle(m, ops.OpListMembers, admin.ListMembers)
	handle(m, ops.OpAddMember, admin.AddMember)
	handle(m, ops.OpRemoveMember, admin.RemoveMember)
	handle(m, ops.OpDemoteMember, admin.DemoteMember)
	handle(m, ops.OpListSettings, admin.ListSettings)
	handle(m, ops.OpSetSetting, admin.SetSetting)
	handle(m, ops.OpResetSetting, admin.ResetSetting)

	handle(m, ops.OpCreateSubmission, subs.CreateSubmission)
	handle(m, ops.OpListSubmissions, subs.ListSubmissions)
	handle(m, ops.OpGetSubmission, subs.GetSubmission)
	handle(m, ops.OpRetrySubmission, subs.RetrySubmission)
	handle(m, ops.OpCancelSubmission, subs.CancelSubmission)
	return m, nil
}

// handle registers op's route on m. The route admits the caller before
// it reads the body, so a caller without op's privilege learns nothing
// from the body's shape, then decodes a Req, calls the operation as the
// verified actor, and answers the result. It panics on an op
// [ops.PrivilegeOf] does not know, which no request can cause.
func handle[Req, Resp any](m *mount, op ops.Op, call func(context.Context, identity.Actor, Req) (Resp, error)) {
	required, ok := ops.PrivilegeOf(op)
	if !ok {
		panic("opshttp: operation " + string(op) + " has no privilege")
	}
	m.routes[op] = func(w http.ResponseWriter, r *http.Request, actor identity.Actor, granted ops.Privilege) {
		if err := admit(op, required, actor, granted); err != nil {
			m.respond(w, op, nil, err)
			return
		}
		var req Req
		if !decode(w, r, &req, bodyCap(op)) {
			return
		}
		resp, err := call(r.Context(), actor, req)
		m.respond(w, op, resp, err)
	}
}

func (m *mount) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, ok := strings.CutPrefix(r.URL.Path, m.base)
	if !ok {
		writeError(w, ops.CodeNotFound, "no such operation")
		return
	}
	actor, granted, err := m.verify(w, r)
	switch {
	case errors.Is(err, ops.CodeUnauthenticated):
		writeError(w, ops.CodeUnauthenticated, "no valid session")
		return
	case err != nil:
		m.log.Error("api session verification failed")
		writeError(w, ops.CodeInternal, "")
		return
	}
	fn, known := m.routes[ops.Op(name)]
	if !known {
		writeError(w, ops.CodeNotFound, "no such operation")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeErrorBody(w, http.StatusMethodNotAllowed, wireError{Code: ops.CodeInvalidRequest, Reason: "method must be POST"})
		return
	}
	if r.URL.RawQuery != "" || r.URL.ForceQuery {
		writeError(w, ops.CodeInvalidRequest, "a query string is not accepted")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, ops.CodeInvalidRequest, "content type must be application/json")
		return
	}
	fn(w, r, actor, granted)
}
