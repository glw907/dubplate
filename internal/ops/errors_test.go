package ops

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
)

// TestCodes pins the literal wire codes: a caller over HTTP reads one of
// these exact strings, so a drifted string here is a wire break no
// compiler catches. Each code is also its error text, and an [*Error]
// built with it, wrapped, matches that code with errors.Is and no other.
func TestCodes(t *testing.T) {
	want := map[Code]string{
		CodeUnavailable:              "unavailable",
		CodeNotFound:                 "not_found",
		CodeConflict:                 "conflict",
		CodeInvalidRequest:           "invalid_request",
		CodeCanceled:                 "canceled",
		CodeForbidden:                "forbidden",
		CodeInternal:                 "internal",
		CodeUnauthenticated:          "unauthenticated",
		CodeReauthenticationRequired: "reauthentication_required",
	}
	for code, literal := range want {
		t.Run(literal, func(t *testing.T) {
			if string(code) != literal || code.Error() != literal {
				t.Errorf("code %q has error text %q, want %q", string(code), code.Error(), literal)
			}
			err := fmt.Errorf("wrapped: %w", NewError(code, "a reason"))
			for other := range want {
				if got := errors.Is(err, other); got != (other == code) {
					t.Errorf("errors.Is(%v, %s) = %v, want %v", err, other, got, other == code)
				}
			}
		})
	}
}

func TestNewError_UnavailableNeedsAReason(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal(`NewError(CodeUnavailable, "") did not panic`)
		}
	}()
	_ = NewError(CodeUnavailable, "")
}

func TestError_ErrorString(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{"with reason", NewError(CodeNotFound, "job abc123 not found"), "not_found: job abc123 not found"},
		{"without reason", NewError(CodeCanceled, ""), "canceled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCodeOf pins the classifier: the first [*Error] in the chain
// wins, else a canceled context answers [CodeCanceled], else
// [CodeInternal] covers everything else, a deadline and nil included.
func TestCodeOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Code
	}{
		{"unavailable, wrapped", fmt.Errorf("read status: %w", NewError(CodeUnavailable, "off")), CodeUnavailable},
		{"not found, wrapped", fmt.Errorf("read status: %w", NewError(CodeNotFound, "gone")), CodeNotFound},
		{"conflict, wrapped", fmt.Errorf("read status: %w", NewError(CodeConflict, "busy")), CodeConflict},
		{"invalid request, wrapped", fmt.Errorf("read status: %w", NewError(CodeInvalidRequest, "bad")), CodeInvalidRequest},
		{"canceled, wrapped", fmt.Errorf("read status: %w", NewError(CodeCanceled, "")), CodeCanceled},
		{"forbidden, wrapped", fmt.Errorf("read status: %w", NewError(CodeForbidden, "")), CodeForbidden},
		{"internal, wrapped", fmt.Errorf("read status: %w", NewError(CodeInternal, "")), CodeInternal},
		{"unauthenticated, wrapped", fmt.Errorf("read status: %w", NewError(CodeUnauthenticated, "")), CodeUnauthenticated},
		{"reauthentication required, wrapped", fmt.Errorf("write: %w", NewError(CodeReauthenticationRequired, "")), CodeReauthenticationRequired},
		{"context.Canceled in the chain, no *Error", fmt.Errorf("write pool: %w", context.Canceled), CodeCanceled},
		{"context.DeadlineExceeded alone", fmt.Errorf("write pool: %w", context.DeadlineExceeded), CodeInternal},
		{"any other error", errors.New("sql: database is closed"), CodeInternal},
		{"nil", nil, CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CodeOf(tt.err); got != tt.want {
				t.Errorf("CodeOf(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

// TestError_WithDetail proves a detail rides an error without changing
// its code or the error it was built from.
func TestError_WithDetail(t *testing.T) {
	base := NewError(CodeConflict, "the last admin")
	got := base.WithDetail(DetailLastAdmin)
	if got.Detail != "last_admin" {
		t.Errorf("Detail = %q, want %q", got.Detail, "last_admin")
	}
	if got.Code != CodeConflict || got.Reason != "the last admin" || !errors.Is(got, CodeConflict) {
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
	base := NewError(CodeInvalidRequest, "the manifest was refused")
	rejects := Rejects{Submission: []reject.Reason{reject.Of(reject.NoAudio)}, Files: []FileReason{}}
	got := base.WithRejects(rejects)
	if got.Rejects == nil || len(got.Rejects.Submission) != 1 || got.Rejects.Submission[0].Code != reject.NoAudio {
		t.Errorf("Rejects = %+v, want the no_audio submission reason", got.Rejects)
	}
	if got.Code != CodeInvalidRequest || got.Reason != "the manifest was refused" || !errors.Is(got, CodeInvalidRequest) {
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
