package ops

import (
	"errors"
	"fmt"
)

// A Code is one of the seam's stable wire error codes, shared by every
// operation rather than numbered per endpoint (OpenSubsonic declined,
// prior art).
type Code string

const (
	CodeUnavailable    Code = "unavailable"
	CodeNotFound       Code = "not_found"
	CodeConflict       Code = "conflict"
	CodeInvalidRequest Code = "invalid_request"
	CodeCanceled       Code = "canceled"
)

// The sentinels 6b-i raises. A caller matches one with [errors.Is].
// [ErrForbidden] and [ErrInternal] are declared where they are first
// raised, not here.
var (
	ErrUnavailable    = errors.New("unavailable")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInvalidRequest = errors.New("invalid request")
	ErrCanceled       = errors.New("canceled")
)

// codes maps each declared sentinel to its stable wire [Code]. A
// sentinel with no entry here is a defect in this file, never a
// caller's.
var codes = map[error]Code{
	ErrUnavailable:    CodeUnavailable,
	ErrNotFound:       CodeNotFound,
	ErrConflict:       CodeConflict,
	ErrInvalidRequest: CodeInvalidRequest,
	ErrCanceled:       CodeCanceled,
}

// An Error wraps one of this package's sentinels with its stable
// [Code] and the reason a caller may report as an answer. Reason is
// required for [ErrUnavailable] (every unavailable answer names why);
// the other sentinels may carry an empty one when the sentinel alone
// says enough.
type Error struct {
	Code   Code
	Reason string
	err    error
}

func newError(sentinel error, reason string) *Error {
	code, ok := codes[sentinel]
	if !ok {
		panic(fmt.Sprintf("ops: newError: %v has no registered code", sentinel))
	}
	return &Error{Code: code, Reason: reason, err: sentinel}
}

// NewUnavailableError builds an [*Error] wrapping [ErrUnavailable]
// that names why the operation cannot run right now. A caller with no
// reason to give has called this in error.
func NewUnavailableError(reason string) *Error {
	if reason == "" {
		panic("ops: NewUnavailableError called with an empty reason")
	}
	return newError(ErrUnavailable, reason)
}

// NewNotFoundError builds an [*Error] wrapping [ErrNotFound].
func NewNotFoundError(reason string) *Error { return newError(ErrNotFound, reason) }

// NewConflictError builds an [*Error] wrapping [ErrConflict].
func NewConflictError(reason string) *Error { return newError(ErrConflict, reason) }

// NewInvalidRequestError builds an [*Error] wrapping [ErrInvalidRequest].
func NewInvalidRequestError(reason string) *Error { return newError(ErrInvalidRequest, reason) }

// NewCanceledError builds an [*Error] wrapping [ErrCanceled].
func NewCanceledError(reason string) *Error { return newError(ErrCanceled, reason) }

// Error implements the error interface, joining e's code and reason.
func (e *Error) Error() string {
	if e.Reason == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

// Unwrap returns the sentinel e wraps, so a caller matches it with
// [errors.Is].
func (e *Error) Unwrap() error { return e.err }
