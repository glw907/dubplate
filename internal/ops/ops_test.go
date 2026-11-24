package ops

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/id"
)

func TestJobState_Terminal(t *testing.T) {
	tests := []struct {
		state JobState
		want  bool
	}{
		{JobQueued, false},
		{JobRunning, false},
		{JobSucceeded, true},
		{JobFailed, true},
		{JobCanceled, true},
		{JobInterrupted, true},
	}
	for _, tt := range tests {
		if got := tt.state.Terminal(); got != tt.want {
			t.Errorf("JobState(%q).Terminal() = %v, want %v", tt.state, got, tt.want)
		}
	}
}

// roundTrip marshals want, unmarshals the result into a fresh T, and
// fails unless it comes back equal: the JSON round trip criterion
// every request and response type must meet.
func roundTrip[T any](t *testing.T, want T) {
	t.Helper()
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal %#v: %v", want, err)
	}
	var got T
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip: got %#v, want %#v", got, want)
	}
}

func TestWireTypesRoundTrip(t *testing.T) {
	fullMember := Member{
		UserID:      id.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		Address:     "ann@example.com",
		DisplayName: "Ann",
		Username:    "ann",
		Admin:       true,
	}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"ListMembersRequest", func(t *testing.T) { roundTrip(t, ListMembersRequest{}) }},
		{"ListMembersResponse", func(t *testing.T) {
			roundTrip(t, ListMembersResponse{Members: []Member{fullMember}})
		}},
		{"AddMemberRequest", func(t *testing.T) {
			roundTrip(t, AddMemberRequest{Address: "ann@example.com", DisplayName: "Ann", Username: "ann"})
		}},
		{"AddMemberResponse", func(t *testing.T) { roundTrip(t, AddMemberResponse{Member: fullMember}) }},
		{"RemoveMemberRequest", func(t *testing.T) { roundTrip(t, RemoveMemberRequest{UserID: fullMember.UserID}) }},
		{"RemoveMemberResponse", func(t *testing.T) { roundTrip(t, RemoveMemberResponse{}) }},
		{"DemoteMemberRequest", func(t *testing.T) { roundTrip(t, DemoteMemberRequest{UserID: fullMember.UserID}) }},
		{"DemoteMemberResponse", func(t *testing.T) { roundTrip(t, DemoteMemberResponse{}) }},
		{"ListSettingsRequest", func(t *testing.T) { roundTrip(t, ListSettingsRequest{}) }},
		{"ListSettingsResponse, store-homed and environment-only", func(t *testing.T) {
			value, source, seed := "2h", "admin", "1h"
			roundTrip(t, ListSettingsResponse{Settings: []SettingEntry{
				{Key: "DUBPLATE_REAUTH_WINDOW", Home: "store", Description: "d", Apply: "next-operation", Sensitivity: "sensitive", Value: &value, Source: &source, SeedValue: &seed},
				{Key: "DUBPLATE_STATE_DIR", Home: "environment", Description: "d", Apply: "restart", Sensitivity: "ordinary"},
			}})
		}},
		{"SetSettingRequest", func(t *testing.T) { roundTrip(t, SetSettingRequest{Key: "K", Value: "v"}) }},
		{"SetSettingResponse", func(t *testing.T) {
			roundTrip(t, SetSettingResponse{Key: "K", Apply: "next-operation", Source: "admin"})
		}},
		{"ResetSettingRequest", func(t *testing.T) { roundTrip(t, ResetSettingRequest{Key: "K"}) }},
		{"ResetSettingResponse", func(t *testing.T) {
			roundTrip(t, ResetSettingResponse{Key: "K", Apply: "restart", Source: "env-seed"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

// TestAddMemberRequest_HasNoAdminField proves the add request type
// cannot carry an admin grant: a body naming one decodes to the same
// request as a body without it.
func TestAddMemberRequest_HasNoAdminField(t *testing.T) {
	var got AddMemberRequest
	body := `{"address":"ann@example.com","display_name":"Ann","username":"ann","admin":true}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := AddMemberRequest{Address: "ann@example.com", DisplayName: "Ann", Username: "ann"}
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

// allowedInternalImports are the internal imports internal/ops may
// carry: the network actor's channel type (Global constraints) and the
// reject set, a leaf that imports nothing, whose reasons the submission
// operations carry on the wire, and the id package, another leaf, whose
// type every wire id carries.
var allowedInternalImports = []string{
	"github.com/glw907/dubplate/internal/id",
	"github.com/glw907/dubplate/internal/identity",
	"github.com/glw907/dubplate/internal/reject",
}

// TestImportsStandardLibraryOnly parses every .go file this package
// declares and fails if any import path's first segment carries a
// dot, the shape every non-standard-library module path takes
// (domain-qualified) and no standard-library path ever does, unless
// the import is exactly one of [allowedInternalImports].
func TestImportsStandardLibraryOnly(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob *.go: %v", err)
	}
	fset := token.NewFileSet()
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imp := range f.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquote import %s: %v", path, imp.Path.Value, err)
			}
			if slices.Contains(allowedInternalImports, importPath) {
				continue
			}
			first, _, _ := strings.Cut(importPath, "/")
			if strings.Contains(first, ".") {
				t.Errorf("%s: imports %s, which internal/ops must not", path, importPath)
			}
		}
	}
}
