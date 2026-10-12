package opshttp

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// ErrUnauthenticated is what a [Verifier] wraps to say the request
// carries no valid session. The mount answers it 401 unauthenticated;
// any other verifier error is the server failing and answers 500.
var ErrUnauthenticated = errors.New("opshttp: no valid session")

// A Verifier admits one request to the mount. It reads whatever
// credential the request carries, may set a refreshed credential on w,
// and answers the actor the call proceeds as and the privilege that
// actor holds. A refused credential wraps [ErrUnauthenticated]. A
// verifier logs its own failures: the mount never writes an error's
// text, which may name a member.
type Verifier func(w http.ResponseWriter, r *http.Request) (ops.Actor, ops.Privilege, error)

// A mount serves the operations under <prefix>/api/v1/ for sessions a
// Verifier admits.
type mount struct {
	base   string
	verify Verifier
	ops    *handler
	log    *slog.Logger
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
// of the request. prefix is the shell's owned prefix, with a leading
// separator and no trailing one. NewMount fails on a nil seam,
// verifier, or logger, and on a prefix of the wrong shape.
func NewMount(jobs ops.Seam, admin ops.AdminSeam, prefix string, verify Verifier, log *slog.Logger) (http.Handler, error) {
	if verify == nil {
		return nil, errors.New("opshttp: verifier is nil")
	}
	if !strings.HasPrefix(prefix, "/") || strings.HasSuffix(prefix, "/") {
		return nil, errors.New("opshttp: prefix must begin with a separator and carry no trailing one")
	}
	h, err := newHandler(jobs, admin, contextAuthorizer{}, identity.ChannelCookie, log)
	if err != nil {
		return nil, err
	}
	return &mount{base: prefix + "/api/v1/", verify: verify, ops: h, log: log}, nil
}

func (m *mount) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, ok := strings.CutPrefix(r.URL.Path, m.base)
	if !ok {
		writeError(w, http.StatusNotFound, ops.CodeNotFound, "no such operation")
		return
	}
	actor, privilege, err := m.verify(w, r)
	switch {
	case errors.Is(err, ErrUnauthenticated):
		writeError(w, statuses[ops.CodeUnauthenticated], ops.CodeUnauthenticated, "no valid session")
		return
	case err != nil:
		m.log.Error("api session verification failed")
		writeError(w, statuses[ops.CodeInternal], ops.CodeInternal, "")
		return
	}
	routed := r.Clone(withAdmitted(r.Context(), actor, privilege))
	routed.URL.Path, routed.URL.RawPath = "/"+name, ""
	m.ops.ServeHTTP(w, routed)
}
