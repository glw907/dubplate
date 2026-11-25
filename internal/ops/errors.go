package ops

import (
	"context"
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/reject"
)

// A Code is one of the operations' stable wire error codes, shared by
// every operation rather than numbered per endpoint. A Code is an
// error, so a caller matches a failure with errors.Is(err, CodeConflict)
// whether the chain holds an [*Error] or the bare code.
type Code string

const (
	CodeUnavailable    Code = "unavailable"
	CodeNotFound       Code = "not_found"
	CodeConflict       Code = "conflict"
	CodeInvalidRequest Code = "invalid_request"
	CodeCanceled       Code = "canceled"
	CodeForbidden      Code = "forbidden"
	CodeInternal       Code = "internal"

	// CodeUnauthenticated answers a request with no valid session.
	CodeUnauthenticated Code = "unauthenticated"

	// CodeReauthenticationRequired answers a gated write from a
	// session anchored outside the re-authentication window.
	CodeReauthenticationRequired Code = "reauthentication_required"
)

// Error returns c's wire text.
func (c Code) Error() string { return string(c) }

// A FileReason is one declared file's reject reason. Path is valid
// UTF-8, with the replacement character for any byte that was not.
type FileReason struct {
	Path   string        `json:"path"`
	Reason reject.Reason `json:"reason"`
}

// Rejects is every reason a refused submission carries: the
// submission-level reasons and one per failing file. Both lists are
// present on the wire, empty when nothing fails at that level.
type Rejects struct {
	Submission []reject.Reason `json:"submission"`
	Files      []FileReason    `json:"files"`
}

// An Error is an operation's failure: its stable [Code] and the reason
// a caller may report as an answer. Detail, when set, is an enumerated
// name finer than Code, which a client may branch on where Reason is
// free text. Rejects, when set, lists the reject reasons behind a
// refused submission.
type Error struct {
	Code    Code
	Reason  string
	Detail  string
	Rejects *Rejects
}

// NewError returns an [*Error] carrying code and reason. Reason may be
// empty when the code alone says enough, except for [CodeUnavailable]:
// every unavailable answer names why, so NewError panics without one.
func NewError(code Code, reason string) *Error {
	if code == CodeUnavailable && reason == "" {
		panic("ops: NewError called for unavailable with an empty reason")
	}
	return &Error{Code: code, Reason: reason}
}

// WithDetail returns a copy of e carrying detail, one of the Detail
// constants. e itself is unchanged.
func (e *Error) WithDetail(detail string) *Error {
	withDetail := *e
	withDetail.Detail = detail
	return &withDetail
}

// WithRejects returns a copy of e carrying r. e itself is unchanged.
func (e *Error) WithRejects(r Rejects) *Error {
	withRejects := *e
	withRejects.Rejects = &r
	return &withRejects
}

// Error implements the error interface, joining e's code and reason.
func (e *Error) Error() string {
	if e.Reason == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

// Unwrap returns e's code, so errors.Is matches e against a [Code].
func (e *Error) Unwrap() error { return e.Code }

// CodeOf returns err's stable wire [Code]: the code of the first
// [*Error] in err's chain; else [CodeCanceled] when the chain holds
// [context.Canceled]; else [CodeInternal], which is also CodeOf(nil).
func CodeOf(err error) Code {
	if opsErr, ok := errors.AsType[*Error](err); ok {
		return opsErr.Code
	}
	if errors.Is(err, context.Canceled) {
		return CodeCanceled
	}
	return CodeInternal
}
