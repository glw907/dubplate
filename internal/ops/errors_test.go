package ops

import (
	"errors"
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
	}
	wantLiterals := map[string]Code{
		"unavailable":     "unavailable",
		"not found":       "not_found",
		"conflict":        "conflict",
		"invalid request": "invalid_request",
		"canceled":        "canceled",
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
