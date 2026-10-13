package ops

import (
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

func TestLocalActor(t *testing.T) {
	got := LocalActor("geoff")
	if got.String() != "local:geoff" {
		t.Errorf("LocalActor(%q).String() = %q, want %q", "geoff", got.String(), "local:geoff")
	}
	if got.Channel() != identity.ChannelCommandLine {
		t.Errorf("LocalActor(%q).Channel() = %v, want %v", "geoff", got.Channel(), identity.ChannelCommandLine)
	}
}

func TestLocalActor_EmptyUsernamePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("LocalActor(\"\") did not panic")
		}
	}()
	LocalActor("")
}

// validIdentity is a cookie-verified login, valid by construction.
func validIdentity() identity.Identity {
	return identity.Identity{
		Address:     "geoff@example.com",
		UserID:      identity.UserID{0x01},
		Principal:   "geoff",
		DisplayName: "Geoff",
		Source:      "magic-link",
		Channel:     identity.ChannelCookie,
	}
}

// testAnchor is the instant every network actor here was anchored.
var testAnchor = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

// TestNetworkActor pins the wire form, the channel, and the two
// read-only accessors for a valid login. The wire form is the
// lowercase hex of the UserID, written out here rather than computed.
func TestNetworkActor(t *testing.T) {
	got := NetworkActor(validIdentity(), testAnchor)
	if want := "member:01000000000000000000000000000000"; got.String() != want {
		t.Errorf("NetworkActor(...).String() = %q, want %q", got.String(), want)
	}
	if got.Channel() != identity.ChannelCookie {
		t.Errorf("NetworkActor(...).Channel() = %v, want %v", got.Channel(), identity.ChannelCookie)
	}
	if want := (identity.UserID{0x01}); got.UserID() != want {
		t.Errorf("NetworkActor(...).UserID() = %x, want %x", got.UserID(), want)
	}
	if !got.Anchor().Equal(testAnchor) {
		t.Errorf("NetworkActor(...).Anchor() = %v, want %v", got.Anchor(), testAnchor)
	}
}

// TestNetworkActor_WireFormFollowsUserID proves the wire form is the
// durable key alone: two logins with one UserID and different
// principals render one form, in lowercase, and a login with another
// UserID renders another.
func TestNetworkActor_WireFormFollowsUserID(t *testing.T) {
	login := func(user identity.UserID, principal string) identity.Identity {
		id := validIdentity()
		id.UserID, id.Principal = user, principal
		return id
	}
	shared := identity.UserID{0xab, 0xcd, 0xef}
	const wantShared = "member:abcdef00000000000000000000000000"
	tests := []struct {
		name string
		id   identity.Identity
		want string
	}{
		{"principal geoff", login(shared, "geoff"), wantShared},
		{"principal renamed", login(shared, "geoffrey"), wantShared},
		{"another member", login(identity.UserID{0x02}, "geoff"), "member:02000000000000000000000000000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NetworkActor(tt.id, testAnchor).String(); got != tt.want {
				t.Errorf("NetworkActor(%+v).String() = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

// TestLocalActor_CarriesNoSession proves the command line has no
// durable id and no anchor to read.
func TestLocalActor_CarriesNoSession(t *testing.T) {
	got := LocalActor("geoff")
	if got.UserID() != (identity.UserID{}) {
		t.Errorf("LocalActor(...).UserID() = %x, want the zero UserID", got.UserID())
	}
	if !got.Anchor().IsZero() {
		t.Errorf("LocalActor(...).Anchor() = %v, want the zero time", got.Anchor())
	}
}

// TestNetworkActor_RefusesInvalidIdentity proves construction refuses
// both the zero identity and a command-line identity: neither is a
// verified login a network caller can assert.
func TestNetworkActor_RefusesInvalidIdentity(t *testing.T) {
	tests := []struct {
		name string
		id   identity.Identity
	}{
		{"zero identity", identity.Identity{}},
		{"command-line identity", func() identity.Identity {
			id := validIdentity()
			id.Channel = identity.ChannelCommandLine
			return id
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("NetworkActor(%+v) did not panic", tt.id)
				}
			}()
			NetworkActor(tt.id, testAnchor)
		})
	}
}

func TestParseHandle(t *testing.T) {
	tests := []struct {
		name    string
		handle  string
		wantErr bool
	}{
		{"32 lowercase hex", "0123456789abcdef0123456789abcdef", false},
		{"empty", "", true},
		{"wrong length, too short", "0123456789abcdef", true},
		{"wrong length, too long", "0123456789abcdef0123456789abcdef00", true},
		{"uppercase", "0123456789ABCDEF0123456789ABCDEF", true},
		{"non-hex", "gggggggggggggggggggggggggggggggg", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHandle(tt.handle)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseHandle(%q) = %q, nil, want an error", tt.handle, got)
				}
				var opsErr *Error
				if !errors.As(err, &opsErr) {
					t.Fatalf("ParseHandle(%q) error %v does not carry *ops.Error", tt.handle, err)
				}
				if opsErr.Code != CodeInvalidRequest {
					t.Errorf("ParseHandle(%q) code = %q, want %q", tt.handle, opsErr.Code, CodeInvalidRequest)
				}
				if !strings.Contains(opsErr.Reason, tt.handle) {
					t.Errorf("ParseHandle(%q) reason %q does not name the handle", tt.handle, opsErr.Reason)
				}
				if !errors.Is(err, ErrInvalidRequest) {
					t.Errorf("ParseHandle(%q) does not wrap ErrInvalidRequest", tt.handle)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseHandle(%q) = %v, want no error", tt.handle, err)
			}
			if string(got) != tt.handle {
				t.Errorf("ParseHandle(%q) = %q, want %q", tt.handle, got, tt.handle)
			}
		})
	}
}

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
	started := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	finished := started.Add(time.Minute)
	fullStatus := JobStatus{
		ID:             "0123456789abcdef0123456789abcdef",
		Kind:           KindBackup,
		State:          JobFailed,
		Phase:          "sending to remote",
		FailureCode:    "rclone-exit-7",
		FailureMessage: "failed to sync",
		QueuedAt:       started.Add(-time.Second),
		StartedAt:      &started,
		UpdatedAt:      finished,
		FinishedAt:     &finished,
	}

	fullMember := Member{
		UserID:      "0123456789abcdef0123456789abcdef",
		Address:     "ann@example.com",
		DisplayName: "Ann",
		Username:    "ann",
		Admin:       true,
	}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"StartBackupRequest", func(t *testing.T) { roundTrip(t, StartBackupRequest{}) }},
		{"StartBackupResponse", func(t *testing.T) {
			roundTrip(t, StartBackupResponse{JobID: "0123456789abcdef0123456789abcdef"})
		}},
		{"JobStatusRequest", func(t *testing.T) {
			roundTrip(t, JobStatusRequest{JobID: "0123456789abcdef0123456789abcdef"})
		}},
		{"JobStatusResponse, minimal", func(t *testing.T) {
			roundTrip(t, JobStatusResponse{Job: JobStatus{ID: "0123456789abcdef0123456789abcdef", Kind: KindBackup, State: JobQueued, QueuedAt: started, UpdatedAt: started}})
		}},
		{"JobStatusResponse, every field set", func(t *testing.T) {
			roundTrip(t, JobStatusResponse{Job: fullStatus})
		}},
		{"CancelJobRequest", func(t *testing.T) {
			roundTrip(t, CancelJobRequest{JobID: "0123456789abcdef0123456789abcdef"})
		}},
		{"CancelJobResponse", func(t *testing.T) {
			roundTrip(t, CancelJobResponse{Job: fullStatus})
		}},
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
				{Name: "DUBPLATE_REAUTH_WINDOW", Home: "store", Description: "d", Apply: "next-operation", Mark: "sensitive", Value: &value, Source: &source, SeedValue: &seed},
				{Name: "DUBPLATE_STATE_DIR", Home: "environment", Description: "d", Apply: "restart", Mark: "ordinary"},
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
		{"CatalogRequest", func(t *testing.T) { roundTrip(t, CatalogRequest{}) }},
		{"CatalogResponse", func(t *testing.T) {
			roundTrip(t, CatalogResponse{Operations: []CatalogEntry{
				{Name: OpStartBackup, Privilege: PrivilegeAdmin, Available: true},
				{Name: OpCatalog, Privilege: PrivilegeMember, Available: false, Reason: "backup integration disabled"},
			}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

// TestStartBackupRequest_IgnoresActorField proves the type half of
// spec criterion 2: an actor field on the wire never reaches the
// decoded request, because the request type carries no such field.
func TestStartBackupRequest_IgnoresActorField(t *testing.T) {
	var got StartBackupRequest
	if err := json.Unmarshal([]byte(`{"actor":"someone"}`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != (StartBackupRequest{}) {
		t.Fatalf("got %#v, want the zero value", got)
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
// operations carry on the wire.
var allowedInternalImports = []string{
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
