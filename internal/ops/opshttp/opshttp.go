package opshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// maxBody is the largest request body read, in bytes, unless bodyCaps
// names the operation. Every other request type is a few short fields,
// so an honest body is a small fraction of this.
const maxBody = 4 << 10

// bodyCaps holds the operations whose honest body outgrows maxBody. A
// create_submission manifest names up to 2000 files, the most the file
// count setting allows, at 4 KiB a path.
var bodyCaps = map[string]int64{
	ops.OpCreateSubmission: 8 << 20,
}

// bodyCap is the largest body op reads, in bytes.
func bodyCap(op string) int64 {
	if limit, ok := bodyCaps[op]; ok {
		return limit
	}
	return maxBody
}

// statuses maps each wire code to the HTTP status that answers it.
var statuses = map[ops.Code]int{
	ops.CodeInvalidRequest: http.StatusBadRequest,
	ops.CodeForbidden:      http.StatusForbidden,
	ops.CodeNotFound:       http.StatusNotFound,
	ops.CodeConflict:       http.StatusConflict,
	ops.CodeUnavailable:    http.StatusServiceUnavailable,
	ops.CodeCanceled:       http.StatusServiceUnavailable,
	ops.CodeInternal:       http.StatusInternalServerError,

	ops.CodeUnauthenticated:          http.StatusUnauthorized,
	ops.CodeReauthenticationRequired: http.StatusForbidden,
}

// checkStatuses fails, naming the first code [ops.Codes] declares that
// table gives no status, so a sentinel added to package ops cannot
// reach the wire without one.
func checkStatuses(table map[ops.Code]int) error {
	for _, c := range ops.Codes() {
		if _, ok := table[c]; !ok {
			return fmt.Errorf("opshttp: wire code %q has no HTTP status", c)
		}
	}
	return nil
}

// A handler serves the twelve operations, each as POST /<name>. It
// dispatches by exact path itself rather than through [http.ServeMux],
// whose method-mismatch and not-found answers are plain text, so that
// every answer here carries the JSON error shape.
type handler struct {
	seam   *authorizingSeam
	log    *slog.Logger
	routes map[string]http.HandlerFunc
}

// newHandler builds the projection of the admin and submission seams,
// admitting each call through auth on a mount whose verified identities
// arrive on channel. It fails on a nil seam, auth, or log, on a channel
// no network mount carries, and on a wire code with no HTTP status.
func newHandler(admin ops.AdminSeam, subs ops.SubmissionSeam, auth authorizer, channel identity.Channel, log *slog.Logger) (*handler, error) {
	if log == nil {
		return nil, errors.New("opshttp: logger is nil")
	}
	switch {
	case admin == nil:
		return nil, errors.New("opshttp: admin seam is nil")
	case subs == nil:
		return nil, errors.New("opshttp: submission seam is nil")
	}
	seam, err := newAuthorizingSeam(auth, channel)
	if err != nil {
		return nil, err
	}
	if err := checkStatuses(statuses); err != nil {
		return nil, err
	}
	h := &handler{seam: seam, log: log}
	h.routes = map[string]http.HandlerFunc{
		ops.OpListMembers:  route(h, ops.OpListMembers, admin.ListMembers),
		ops.OpAddMember:    route(h, ops.OpAddMember, admin.AddMember),
		ops.OpRemoveMember: route(h, ops.OpRemoveMember, admin.RemoveMember),
		ops.OpDemoteMember: route(h, ops.OpDemoteMember, admin.DemoteMember),
		ops.OpListSettings: route(h, ops.OpListSettings, admin.ListSettings),
		ops.OpSetSetting:   route(h, ops.OpSetSetting, admin.SetSetting),
		ops.OpResetSetting: route(h, ops.OpResetSetting, admin.ResetSetting),

		ops.OpCreateSubmission: route(h, ops.OpCreateSubmission, subs.CreateSubmission),
		ops.OpListSubmissions:  route(h, ops.OpListSubmissions, subs.ListSubmissions),
		ops.OpGetSubmission:    route(h, ops.OpGetSubmission, subs.GetSubmission),
		ops.OpRetrySubmission:  route(h, ops.OpRetrySubmission, subs.RetrySubmission),
		ops.OpCancelSubmission: route(h, ops.OpCancelSubmission, subs.CancelSubmission),
	}
	return h, nil
}

// route builds the handler of an operation whose request needs no check
// beyond its decoding. It admits the caller before it reads the body, so
// a caller without the operation's privilege learns nothing from the
// body's shape, then decodes a Req, calls the operation as the admitted
// actor, and answers the result.
func route[Req, Resp any](h *handler, op string, call func(context.Context, identity.Actor, Req) (Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, err := h.seam.admit(r.Context(), op)
		if err != nil {
			h.respond(w, op, nil, err)
			return
		}
		var req Req
		if !decode(w, r, &req, bodyCap(op)) {
			return
		}
		resp, err := call(r.Context(), actor, req)
		h.respond(w, op, resp, err)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, ok := strings.CutPrefix(r.URL.Path, "/")
	route, known := h.routes[name]
	if !ok || !known {
		writeError(w, http.StatusNotFound, ops.CodeNotFound, "no such operation")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, ops.CodeInvalidRequest, "method must be POST")
		return
	}
	if r.URL.RawQuery != "" || r.URL.ForceQuery {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "a query string is not accepted")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "content type must be application/json")
		return
	}
	route(w, r)
}

// decode reads r's body, at most limit bytes, into v, which must hold
// exactly one JSON object, and reports whether it did; on a refusal it has already
// answered w. Unknown keys are ignored, so a body naming an actor or a
// privilege changes nothing: neither is a field of any request type.
func decode(w http.ResponseWriter, r *http.Request, v any, limit int64) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if err != nil {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, fmt.Sprintf("request body exceeds %d bytes or could not be read", limit))
		return false
	}
	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "request body is empty")
		return false
	}
	// Decoding null into a struct succeeds and leaves the zero request,
	// which for a list operation is a valid one, so the object is required
	// before any decode.
	if tok, err := json.NewDecoder(bytes.NewReader(body)).Token(); err != nil || tok != json.Delim('{') {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "request body is not a JSON object")
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "request body does not match the operation's request")
		return false
	}
	if _, err := dec.Token(); err != io.EOF {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "request body carries data after its object")
		return false
	}
	return true
}

// respond answers an operation's result. An error's body carries its
// wire code and, unless that code is internal, the reason of the first
// [*ops.Error] in its chain, never the chain's own text: a wrapper
// around a coded error may name a path, and an internal error may be
// anything. A code statuses lacks, such as a zero [ops.Error]'s, answers
// as internal. The log names the operation with the whole error, except
// a forbidden one, which it names by its cause from [denialReason]
// alone: a denial's chain may carry the caller's identity.
func (h *handler) respond(w http.ResponseWriter, op string, resp any, err error) {
	if err != nil {
		code := ops.CodeOf(err)
		status, mapped := statuses[code]
		switch {
		case !mapped:
			h.log.Error("ops error code has no HTTP status", "op", op, "code", code, "err", err)
			writeError(w, statuses[ops.CodeInternal], ops.CodeInternal, "")
			return
		case code == ops.CodeInternal:
			h.log.Error("ops operation failed", "op", op, "err", err)
			writeError(w, status, code, "")
			return
		case code == ops.CodeForbidden:
			h.log.Info("ops operation refused", "op", op, "code", code, "reason", denialReason(err))
		default:
			h.log.Info("ops operation refused", "op", op, "code", code, "err", err)
		}
		body := errorDetail{Code: code}
		if opsErr, ok := errors.AsType[*ops.Error](err); ok {
			body.Reason, body.Detail, body.Rejects = opsErr.Reason, opsErr.Detail, opsErr.Rejects
		}
		writeErrorBody(w, status, body)
		return
	}
	data, err := json.Marshal(resp)
	if err != nil {
		h.log.Error("ops response encoding failed", "op", op, "err", err)
		writeError(w, statuses[ops.CodeInternal], ops.CodeInternal, "")
		return
	}
	writeBody(w, http.StatusOK, data)
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code   ops.Code `json:"code"`
	Reason string   `json:"reason,omitempty"`
	Detail string   `json:"detail,omitempty"`
	// Rejects is set on a refused submission.
	Rejects *ops.Rejects `json:"rejects,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code ops.Code, reason string) {
	writeErrorBody(w, status, errorDetail{Code: code, Reason: reason})
}

func writeErrorBody(w http.ResponseWriter, status int, e errorDetail) {
	// A struct of three strings always marshals.
	data, _ := json.Marshal(errorBody{Error: e})
	writeBody(w, status, data)
}

// writeBody sends every answer, so each one, a refusal included, is
// marked uncacheable: an operation's answer describes state at the
// moment it ran.
func writeBody(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	// The status is already sent, so a failed write has no answer left
	// to give; the client sees a truncated body.
	_, _ = w.Write(data)
}
