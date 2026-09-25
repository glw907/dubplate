package ops

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
)

// TestSentinelCodes pins R-6bi5's literal wire codes against the
// package's declared sentinels: a caller over HTTP reads one of these
// exact strings, so a drifted string here is a wire break no compiler
// catches.
func TestSentinelCodes(t *testing.T) {
	tests := []struct {
		name     string
		buildErr func(reason string) *Error
		wantCode Code
		sentinel error
	}{
		{"unavailable", NewUnavailableError, CodeUnavailable, ErrUnavailable},
		{"not found", NewNotFoundError, CodeNotFound, ErrNotFound},
		{"conflict", NewConflictError, CodeConflict, ErrConflict},
		{"invalid request", NewInvalidRequestError, CodeInvalidRequest, ErrInvalidRequest},
		{"canceled", NewCanceledError, CodeCanceled, ErrCanceled},
		{"forbidden", NewForbiddenError, CodeForbidden, ErrForbidden},
		{"internal", NewInternalError, CodeInternal, ErrInternal},
	}
	wantLiterals := map[string]Code{
		"unavailable":     "unavailable",
		"not found":       "not_found",
		"conflict":        "conflict",
		"invalid request": "invalid_request",
		"canceled":        "canceled",
		"forbidden":       "forbidden",
		"internal":        "internal",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantCode != wantLiterals[tt.name] {
				t.Fatalf("test table's own wantCode %q disagrees with the pinned literal %q", tt.wantCode, wantLiterals[tt.name])
			}
			got := tt.buildErr("a reason")
			if got.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", got.Code, tt.wantCode)
			}
			if !errors.Is(got, tt.sentinel) {
				t.Errorf("wraps %v, want it to wrap %v", got, tt.sentinel)
			}
		})
	}
}

// TestNewError_UnregisteredSentinelPanics proves newError refuses any
// sentinel absent from codes, the mechanism that would catch a new
// sentinel added to the vocabulary without a codes entry: a test that
// only pinned the five already registered above could not tell that
// mechanism from one that always returns the same code.
func TestNewError_UnregisteredSentinelPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("newError did not panic for an unregistered sentinel")
		}
	}()
	unregistered := errors.New("ops test: not a declared sentinel")
	_ = newError(unregistered, "reason")
}

func TestNewUnavailableError_EmptyReasonPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewUnavailableError(\"\") did not panic")
		}
	}()
	_ = NewUnavailableError("")
}

func TestError_ErrorString(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{"with reason", NewNotFoundError("job abc123 not found"), "not_found: job abc123 not found"},
		{"without reason", NewCanceledError(""), "canceled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCodeOf pins R-6bii5's classifier: the first [*Error] in the
// chain wins, else a canceled context answers [CodeCanceled], else
// [CodeInternal] covers everything else, a deadline included.
func TestCodeOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Code
	}{
		{"unavailable, wrapped", fmt.Errorf("read status: %w", NewUnavailableError("off")), CodeUnavailable},
		{"not found, wrapped", fmt.Errorf("read status: %w", NewNotFoundError("gone")), CodeNotFound},
		{"conflict, wrapped", fmt.Errorf("read status: %w", NewConflictError("busy")), CodeConflict},
		{"invalid request, wrapped", fmt.Errorf("read status: %w", NewInvalidRequestError("bad")), CodeInvalidRequest},
		{"canceled sentinel, wrapped", fmt.Errorf("read status: %w", NewCanceledError("")), CodeCanceled},
		{"forbidden, wrapped", fmt.Errorf("read status: %w", NewForbiddenError("")), CodeForbidden},
		{"internal sentinel, wrapped", fmt.Errorf("read status: %w", NewInternalError("")), CodeInternal},
		{"context.Canceled in the chain, no *Error", fmt.Errorf("write pool: %w", context.Canceled), CodeCanceled},
		{"context.DeadlineExceeded alone", fmt.Errorf("write pool: %w", context.DeadlineExceeded), CodeInternal},
		{"any other error", errors.New("sql: database is closed"), CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CodeOf(tt.err); got != tt.want {
				t.Errorf("CodeOf(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

// TestCodes proves the coverage guard's source: exactly the seven
// literal wire codes, so a sentinel added to codes without a
// corresponding contract case would grow this set past seven.
func TestCodes(t *testing.T) {
	want := []Code{
		CodeUnavailable, CodeNotFound, CodeConflict, CodeInvalidRequest,
		CodeCanceled, CodeForbidden, CodeInternal,
	}
	got := Codes()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Codes() = %v, want %v", got, want)
	}
}
