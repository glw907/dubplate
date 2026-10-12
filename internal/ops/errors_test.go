package ops

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
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
		{"unauthenticated", NewUnauthenticatedError, CodeUnauthenticated, ErrUnauthenticated},
		{"reauthentication required", NewReauthenticationRequiredError, CodeReauthenticationRequired, ErrReauthenticationRequired},
	}
	wantLiterals := map[string]Code{
		"unauthenticated":           "unauthenticated",
		"reauthentication required": "reauthentication_required",
		"unavailable":               "unavailable",
		"not found":                 "not_found",
		"conflict":                  "conflict",
		"invalid request":           "invalid_request",
		"canceled":                  "canceled",
		"forbidden":                 "forbidden",
		"internal":                  "internal",
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
		{"unauthenticated, wrapped", fmt.Errorf("read status: %w", NewUnauthenticatedError("")), CodeUnauthenticated},
		{"reauthentication required, wrapped", fmt.Errorf("write: %w", NewReauthenticationRequiredError("")), CodeReauthenticationRequired},
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

// TestCodes proves the coverage guard's source: exactly the nine
// literal wire codes, so a sentinel added to codes without a
// corresponding contract case would grow this set past nine.
func TestCodes(t *testing.T) {
	want := []Code{
		"unavailable", "not_found", "conflict", "invalid_request",
		"canceled", "forbidden", "internal",
		"unauthenticated", "reauthentication_required",
	}
	got := Codes()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Codes() = %v, want %v", got, want)
	}
}

// TestError_WithDetail proves a detail rides an error without changing
// its code, its sentinel, or the error it was built from.
func TestError_WithDetail(t *testing.T) {
	base := NewConflictError("the last admin")
	got := base.WithDetail(DetailLastAdmin)
	if got.Detail != "last_admin" {
		t.Errorf("Detail = %q, want %q", got.Detail, "last_admin")
	}
	if got.Code != CodeConflict || got.Reason != "the last admin" || !errors.Is(got, ErrConflict) {
		t.Errorf("WithDetail changed the error: %+v", got)
	}
	if base.Detail != "" {
		t.Errorf("WithDetail wrote through to its receiver: Detail = %q", base.Detail)
	}
	wrapped := fmt.Errorf("add member: %w", got)
	if e, ok := errors.AsType[*Error](wrapped); !ok || e.Detail != "last_admin" {
		t.Errorf("AsType over a wrapped detail = %+v, %v", e, ok)
	}
}

func TestError_WithRejects(t *testing.T) {
	base := NewInvalidRequestError("the manifest was refused")
	rejects := Rejects{Submission: []reject.Reason{reject.Of(reject.NoAudio)}, Files: []FileReject{}}
	got := base.WithRejects(rejects)
	if got.Rejects == nil || len(got.Rejects.Submission) != 1 || got.Rejects.Submission[0].Code != reject.NoAudio {
		t.Errorf("Rejects = %+v, want the no-audio submission reason", got.Rejects)
	}
	if got.Code != CodeInvalidRequest || got.Reason != "the manifest was refused" || !errors.Is(got, ErrInvalidRequest) {
		t.Errorf("WithRejects changed the error: %+v", got)
	}
	if base.Rejects != nil {
		t.Errorf("WithRejects wrote through to its receiver: Rejects = %+v", base.Rejects)
	}
	wrapped := fmt.Errorf("create: %w", got)
	if e, ok := errors.AsType[*Error](wrapped); !ok || e.Rejects == nil {
		t.Errorf("AsType over wrapped rejects = %+v, %v", e, ok)
	}
}

// TestDetailConstants pins the thirteen wire details as literals read from
// the source file the parity and deck gates parse: exactly these
// names, each with exactly this value.
func TestDetailConstants(t *testing.T) {
	want := map[string]string{
		"DetailDuplicateAddress":   "duplicate_address",
		"DetailDuplicateUsername":  "duplicate_username",
		"DetailInvalidAddress":     "invalid_address",
		"DetailInvalidDisplayName": "invalid_display_name",
		"DetailBlankDisplayName":   "blank_display_name",
		"DetailInvalidUsername":    "invalid_username",
		"DetailMissingUsername":    "missing_username",
		"DetailLastAdmin":          "last_admin",
		"DetailEnvironmentSets":    "environment_sets",
		"DetailBusy":               "busy",
		"DetailStoreURL":           "store_url",
		"DetailExtra":              "extra",
		"DetailManifest":           "manifest",
	}
	f, err := parser.ParseFile(token.NewFileSet(), "details.go", nil, 0)
	if err != nil {
		t.Fatalf("parse details.go: %v", err)
	}
	got := make(map[string]string)
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok {
					t.Fatalf("%s is not a literal", name.Name)
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatalf("%s: %v", name.Name, err)
				}
				got[name.Name] = value
			}
		}
	}
	if !maps.Equal(got, want) {
		t.Errorf("details.go declares %v, want %v", got, want)
	}
}
