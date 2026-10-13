package intake

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/reject"
	"github.com/tus/tusd/v2/pkg/handler"
)

// ErrUnauthenticated is what a [Verifier] wraps to say the request
// carries no valid session. The tus mount answers it 401; any other
// verifier error is the server failing and answers 500.
var ErrUnauthenticated = errors.New("intake: no valid session")

// A Verifier admits one request to the tus mount: it reads the
// request's session, may set a refreshed credential on w, and answers
// the member the request acts as. The member comes from here alone,
// never from a header, a tus metadata key, or a path. A verifier logs
// its own failures: the mount never writes an error's text.
type Verifier func(w http.ResponseWriter, r *http.Request) (identity.UserID, error)

// Every refusal on the tus path is in tusd's own format, a status and a
// text/plain body of the code, a colon, and a reason, so a client reads
// one shape whichever layer refused. Each definitive refusal is a 4xx
// tus-js-client does not retry; a transient one is a 503 it does.
var (
	errTusUnauthenticated = handler.NewError(string(ops.CodeUnauthenticated), "no valid session", http.StatusUnauthorized)
	errTusInternal        = handler.NewError(string(ops.CodeInternal), "the server failed", http.StatusInternalServerError)
	errTusMethodOverride  = handler.NewError(string(ops.CodeInvalidRequest), "X-HTTP-Method-Override is not accepted", http.StatusBadRequest)
	errTusConcat          = handler.NewError(string(ops.CodeInvalidRequest), "Upload-Concat is not accepted", http.StatusBadRequest)
	errTusNoFile          = handler.NewError(string(ops.CodeInvalidRequest), "the metadata names no file still to send in an open upload of yours", http.StatusBadRequest)
	errTusSizeDiffers     = handler.NewError(string(ops.CodeInvalidRequest), "Upload-Length differs from the size the upload declared for this file", http.StatusBadRequest)
	errTusUploadExists    = handler.NewError(string(ops.CodeInvalidRequest), "this file already has an upload; resume it at its URL", http.StatusBadRequest)
	errTusClosed          = handler.NewError(string(ops.CodeConflict), "the upload this file belongs to has closed", http.StatusForbidden)
	errTusNoSpace         = handler.NewError(string(reject.NoSpace), reject.Of(reject.NoSpace).Friendly, http.StatusForbidden)
	errTusCouldNotCheck   = handler.NewError(string(reject.CouldNotCheck), reject.Of(reject.CouldNotCheck).Friendly, http.StatusServiceUnavailable)
)

// memberKey is the context key the wrapper places the verified member
// under. tusd's callbacks read it from HookEvent.Context, which keeps
// the request context's values.
type memberKey struct{}

// ServeHTTP authorizes the request, then hands tusd the path below
// <prefix>/tus with the member on its context. tusd offers no hook on
// PATCH or HEAD, so nothing reaches it unless every step here passes,
// in order:
//
//  1. a method override is refused, since tusd would let a POST carrying
//     one become a DELETE;
//  2. Upload-Concat is refused, since tusd with no concater ignores it;
//  3. the verifier must admit a member;
//  4. a path naming an upload must parse as a file id the member
//     uploads, in a submission still open, and still uploading for a
//     PATCH; an unknown id, another member's, and a closed one all
//     answer one 404;
//  5. free space must be at or above the floor before a POST or PATCH
//     body is read.
//
// The id is computed from exactly the path tusd receives, as tusd
// computes it, so a path cannot authorize one upload and act on
// another. The wrapper wraps no ResponseWriter: tusd extends the
// connection's deadlines through http.ResponseController, which needs
// the writer net/http handed the shell or one that unwraps to it.
func (m *tusMount) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-HTTP-Method-Override") != "" {
		refuse(w, errTusMethodOverride)
		return
	}
	if r.Header.Get("Upload-Concat") != "" {
		refuse(w, errTusConcat)
		return
	}

	member, err := m.verify(w, r)
	switch {
	case errors.Is(err, ErrUnauthenticated):
		refuse(w, errTusUnauthenticated)
		return
	case err != nil:
		refuse(w, errTusInternal)
		return
	}

	rest, ok := strings.CutPrefix(r.URL.Path, m.base)
	if !ok {
		refuse(w, handler.ErrNotFound)
		return
	}
	if id := strings.Trim(rest, "/"); id != "" {
		if err := m.authorize(r.Context(), member, id, r.Method); err != nil {
			refuse(w, err)
			return
		}
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPatch {
		if err := m.checkFloor(r.Context()); err != nil {
			refuse(w, err)
			return
		}
	}

	routed := r.Clone(context.WithValue(r.Context(), memberKey{}, member))
	routed.URL.Path, routed.URL.RawPath = rest, ""
	m.tusd.ServeHTTP(w, routed)
}

// authorize returns the refusal for a request naming upload id, or nil
// when member may make it. PATCH writes, so it needs the submission
// still uploading; any other method needs it open. Every refusal but a
// failed read is the one 404, so no answer tells another member's
// upload from one that does not exist.
func (m *tusMount) authorize(ctx context.Context, member identity.UserID, id, method string) error {
	file, ok := parseID(id)
	if !ok {
		return handler.ErrNotFound
	}
	var current state
	err := m.subs.store.QueryRow(ctx, `SELECT s.state FROM submission_files f JOIN submissions s ON s.id = f.submission_id
		WHERE f.id = ? AND s.uploader = ? AND s.`+openSubmission, file, member[:]).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return handler.ErrNotFound
	}
	if err != nil {
		m.subs.log.Error("tus request could not read its upload", "error", err)
		return errTusCouldNotCheck
	}
	if method == http.MethodPatch && current != stateUploading {
		return handler.ErrNotFound
	}
	return nil
}

// checkFloor returns the refusal for a write while free space on
// staging is under the floor in force, or nil. A read that fails
// refuses too: a check that cannot run is never a pass.
func (m *tusMount) checkFloor(ctx context.Context) error {
	floor, err := m.floor(ctx)
	if err != nil {
		m.subs.log.Error("tus request could not read the free-space floor", "error", err)
		return errTusCouldNotCheck
	}
	free := m.subs.readFree()
	if free.err != nil {
		m.subs.log.Error("tus request could not read free space", "error", free.err)
		return errTusCouldNotCheck
	}
	if free.bytes < uint64(floor) {
		return errTusNoSpace
	}
	return nil
}

// refuse writes err as tusd writes an error, except that an error that
// is not a [handler.Error] answers errTusInternal rather than its text.
func refuse(w http.ResponseWriter, err error) {
	var e handler.Error
	if !errors.As(err, &e) {
		e = errTusInternal
	}
	for key, value := range e.HTTPResponse.Header {
		w.Header().Set(key, value)
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(e.HTTPResponse.Body)))
	w.WriteHeader(e.HTTPResponse.StatusCode)
	_, _ = io.WriteString(w, e.HTTPResponse.Body)
}
