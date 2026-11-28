package opshttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/ops"
)

// defaultBodyCap is the largest request body read, in bytes, unless
// bodyCaps names the operation. Every other request type is a few short
// fields, so an honest body is a small fraction of this.
const defaultBodyCap = 4 << 10

// bodyCaps holds the operations whose honest body outgrows
// defaultBodyCap. A create_submission manifest names up to 2000 files,
// the most the file count setting allows, at 4 KiB a path.
var bodyCaps = map[ops.Op]int64{
	ops.OpCreateSubmission: 8 << 20,
}

// bodyCap is the largest body op reads, in bytes.
func bodyCap(op ops.Op) int64 {
	if limit, ok := bodyCaps[op]; ok {
		return limit
	}
	return defaultBodyCap
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

// decode reads r's body, at most limit bytes, into v, which must hold
// exactly one JSON object, and reports whether it did; on a refusal it
// has already answered w. Unknown keys are ignored, so a body naming an
// actor or a privilege changes nothing: neither is a field of any
// request type.
func decode(w http.ResponseWriter, r *http.Request, v any, limit int64) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if err != nil {
		writeError(w, ops.CodeInvalidRequest, fmt.Sprintf("request body exceeds %d bytes or could not be read", limit))
		return false
	}
	if len(body) == 0 {
		writeError(w, ops.CodeInvalidRequest, "request body is empty")
		return false
	}
	// Decoding null into a struct succeeds and leaves the zero request,
	// which for a list operation is a valid one, so the object is required
	// before any decode.
	if tok, err := json.NewDecoder(bytes.NewReader(body)).Token(); err != nil || tok != json.Delim('{') {
		writeError(w, ops.CodeInvalidRequest, "request body is not a JSON object")
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	if err := dec.Decode(v); err != nil {
		// The decoder returns an id field's UnmarshalText error as it
		// came, so the sentinel survives to name the refusal.
		reason := "request body does not match the operation's request"
		if errors.Is(err, id.ErrSyntax) {
			reason = "request body names a malformed id: " + id.ErrSyntax.Error()
		}
		writeError(w, ops.CodeInvalidRequest, reason)
		return false
	}
	if _, err := dec.Token(); err != io.EOF {
		writeError(w, ops.CodeInvalidRequest, "request body carries data after its object")
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
func (m *mount) respond(w http.ResponseWriter, op ops.Op, resp any, err error) {
	if err != nil {
		code := ops.CodeOf(err)
		status, mapped := statuses[code]
		switch {
		case !mapped:
			m.log.Error("ops error code has no HTTP status", "op", op, "code", code, "err", err)
			writeError(w, ops.CodeInternal, "")
			return
		case code == ops.CodeInternal:
			m.log.Error("ops operation failed", "op", op, "err", err)
			writeError(w, code, "")
			return
		case code == ops.CodeForbidden:
			m.log.Info("ops operation refused", "op", op, "code", code, "reason", denialReason(err))
		default:
			m.log.Info("ops operation refused", "op", op, "code", code, "err", err)
		}
		body := wireError{Code: code}
		if opsErr, ok := errors.AsType[*ops.Error](err); ok {
			body.Reason, body.Detail, body.Rejects = opsErr.Reason, opsErr.Detail, opsErr.Rejects
		}
		writeErrorBody(w, status, body)
		return
	}
	data, err := json.Marshal(resp)
	if err != nil {
		m.log.Error("ops response encoding failed", "op", op, "err", err)
		writeError(w, ops.CodeInternal, "")
		return
	}
	writeBody(w, http.StatusOK, data)
}

type errorBody struct {
	Error wireError `json:"error"`
}

type wireError struct {
	Code   ops.Code   `json:"code"`
	Reason string     `json:"reason,omitempty"`
	Detail ops.Detail `json:"detail,omitempty"`
	// Rejects is set on a refused submission.
	Rejects *ops.Rejects `json:"rejects,omitempty"`
}

// writeError answers code with the status statuses maps it to.
func writeError(w http.ResponseWriter, code ops.Code, reason string) {
	writeErrorBody(w, statuses[code], wireError{Code: code, Reason: reason})
}

func writeErrorBody(w http.ResponseWriter, status int, e wireError) {
	// errorBody holds only strings and slices of plain structs, so
	// Marshal cannot fail.
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
