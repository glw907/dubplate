package opshttp

import (
	"bytes"
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

// maxBody is the largest request body read, in bytes. Every request
// type is an empty object or a single job handle, so an honest body
// is a small fraction of this.
const maxBody = 4 << 10

// statuses maps each wire code to the HTTP status that answers it.
var statuses = map[ops.Code]int{
	ops.CodeInvalidRequest: http.StatusBadRequest,
	ops.CodeForbidden:      http.StatusForbidden,
	ops.CodeNotFound:       http.StatusNotFound,
	ops.CodeConflict:       http.StatusConflict,
	ops.CodeUnavailable:    http.StatusServiceUnavailable,
	ops.CodeCanceled:       http.StatusServiceUnavailable,
	ops.CodeInternal:       http.StatusInternalServerError,
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

// A handler serves the four operations, each as POST /<name>. It
// dispatches by exact path itself rather than through [http.ServeMux],
// whose method-mismatch and not-found answers are plain text, so that
// every answer here carries the JSON error shape.
type handler struct {
	seam   *authorizingSeam
	log    *slog.Logger
	routes map[string]http.HandlerFunc
}

// newHandler builds the projection of next, admitting each call
// through auth on a mount whose verified identities arrive on channel.
// It fails on a nil next, auth, or log, on a channel no network mount
// carries, and on a wire code with no HTTP status.
func newHandler(next ops.Seam, auth authorizer, channel identity.Channel, log *slog.Logger) (*handler, error) {
	if log == nil {
		return nil, errors.New("opshttp: logger is nil")
	}
	seam, err := newAuthorizingSeam(next, auth, channel)
	if err != nil {
		return nil, err
	}
	if err := checkStatuses(statuses); err != nil {
		return nil, err
	}
	h := &handler{seam: seam, log: log}
	h.routes = map[string]http.HandlerFunc{
		ops.OpStartBackup: h.startBackup,
		ops.OpJobStatus:   h.jobStatus,
		ops.OpCancelJob:   h.cancelJob,
		ops.OpCatalog:     h.catalog,
	}
	return h, nil
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

func (h *handler) startBackup(w http.ResponseWriter, r *http.Request) {
	var req ops.StartBackupRequest
	if !decode(w, r, &req) {
		return
	}
	resp, err := h.seam.StartBackup(r.Context(), ops.Actor{}, req)
	h.respond(w, ops.OpStartBackup, resp, err)
}

func (h *handler) jobStatus(w http.ResponseWriter, r *http.Request) {
	var req ops.JobStatusRequest
	if !decode(w, r, &req) || !validHandle(w, req.JobID) {
		return
	}
	resp, err := h.seam.JobStatus(r.Context(), ops.Actor{}, req)
	h.respond(w, ops.OpJobStatus, resp, err)
}

func (h *handler) cancelJob(w http.ResponseWriter, r *http.Request) {
	var req ops.CancelJobRequest
	if !decode(w, r, &req) || !validHandle(w, req.JobID) {
		return
	}
	resp, err := h.seam.CancelJob(r.Context(), ops.Actor{}, req)
	h.respond(w, ops.OpCancelJob, resp, err)
}

func (h *handler) catalog(w http.ResponseWriter, r *http.Request) {
	var req ops.CatalogRequest
	if !decode(w, r, &req) {
		return
	}
	resp, err := h.seam.Catalog(r.Context(), ops.Actor{}, req)
	h.respond(w, ops.OpCatalog, resp, err)
}

// decode reads r's body into v, which must hold exactly one JSON
// object, and reports whether it did; on a refusal it has already
// answered w. Unknown keys are ignored, so a body naming an actor or a
// privilege changes nothing: neither is a field of any request type.
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, fmt.Sprintf("request body exceeds %d bytes or could not be read", maxBody))
		return false
	}
	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, "request body is empty")
		return false
	}
	// Decoding null into a struct succeeds and leaves the zero request,
	// which for start_backup is a valid one, so the object is required
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

// validHandle reports whether id parses as an [ops.Handle]; on a
// refusal it has already answered w.
func validHandle(w http.ResponseWriter, id ops.Handle) bool {
	if _, err := ops.ParseHandle(string(id)); err != nil {
		reason := "job handle is malformed"
		if opsErr, ok := errors.AsType[*ops.Error](err); ok {
			reason = opsErr.Reason
		}
		writeError(w, http.StatusBadRequest, ops.CodeInvalidRequest, reason)
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
		var reason string
		if opsErr, ok := errors.AsType[*ops.Error](err); ok {
			reason = opsErr.Reason
		}
		writeError(w, status, code, reason)
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
}

func writeError(w http.ResponseWriter, status int, code ops.Code, reason string) {
	// A struct of two strings always marshals.
	data, _ := json.Marshal(errorBody{Error: errorDetail{Code: code, Reason: reason}})
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
