package ops

import (
	"context"
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
	CodeForbidden      Code = "forbidden"
	CodeInternal       Code = "internal"
)

// The sentinels this package raises. A caller matches one with
// [errors.Is].
var (
	ErrUnavailable    = errors.New("unavailable")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInvalidRequest = errors.New("invalid request")
	ErrCanceled       = errors.New("canceled")
	ErrForbidden      = errors.New("forbidden")
	ErrInternal       = errors.New("internal")
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
	ErrForbidden:      CodeForbidden,
	ErrInternal:       CodeInternal,
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

// NewForbiddenError builds an [*Error] wrapping [ErrForbidden].
func NewForbiddenError(reason string) *Error { return newError(ErrForbidden, reason) }

// NewInternalError builds an [*Error] wrapping [ErrInternal].
func NewInternalError(reason string) *Error { return newError(ErrInternal, reason) }

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

// CodeOf returns err's stable wire [Code]: the code of the first
// [*Error] in err's chain ([errors.AsType]); else [CodeCanceled] when
// the chain holds [context.Canceled]; else [CodeInternal]. A
// projection calls CodeOf only for a non-nil operation error, so a
// success answers with no call to CodeOf at all.
func CodeOf(err error) Code {
	if opsErr, ok := errors.AsType[*Error](err); ok {
		return opsErr.Code
	}
	if errors.Is(err, context.Canceled) {
		return CodeCanceled
	}
	return CodeInternal
}

// Codes returns every wire [Code] this package's sentinel table
// holds, in no fixed order. A coverage guard that iterates Codes
// catches a sentinel added to codes with no matching case; a guard
// over a list a test hard-codes would only prove the list matches
// itself.
func Codes() []Code {
	out := make([]Code, 0, len(codes))
	for _, c := range codes {
		out = append(out, c)
	}
	return out
}
