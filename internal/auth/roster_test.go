package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// writeRosterFile writes content to a path under t.TempDir() named
// "roster.json" and returns the path.
func writeRosterFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "roster.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture roster: %v", err)
	}
	return path
}

// mustList returns r's members, failing the test when the roster does
// not load.
func mustList(t *testing.T, r *Roster) []Member {
	t.Helper()
	members, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	return members
}

// TestOpenRoster_LoadRefusals proves the four fail-closed refusals
// (R2), each by its own sentinel and each naming the path: missing,
// unreadable, unparseable, and a record failing validation. Named
// mutation 1 (return an empty roster and a nil error for a missing
// file) fails the "missing file" row here.
func TestOpenRoster_LoadRefusals(t *testing.T) {
	tests := []struct {
		name    string
		path    func(t *testing.T) string
		wantErr error
	}{
		{
			name: "missing file",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist.json")
			},
			wantErr: ErrRosterMissing,
		},
		{
			name: "unreadable file",
			path: func(t *testing.T) string {
				if os.Geteuid() == 0 {
					t.Skip("running as root: permission bits do not block reads")
				}
				path := writeRosterFile(t, `{"schema_version":1,"members":[]}`)
				if err := os.Chmod(path, 0o000); err != nil {
					t.Fatalf("Chmod: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
				return path
			},
			wantErr: ErrRosterUnreadable,
		},
		{
			name: "unparseable JSON",
			path: func(t *testing.T) string {
				return writeRosterFile(t, `{not valid json`)
			},
			wantErr: ErrRosterUnparseable,
		},
		{
			name: "record failing validation",
			path: func(t *testing.T) string {
				return writeRosterFile(t, `{"schema_version":1,"members":[{"address":"not-an-address","display_name":"A"}]}`)
			},
			wantErr: ErrRosterRecordInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path(t)
			_, err := OpenRoster(path)
			if err == nil {
				t.Fatalf("OpenRoster(%q) error = nil, want %v", path, tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("OpenRoster(%q) error = %v, want it to wrap %v", path, err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), path) {
				t.Errorf("OpenRoster(%q) error = %q, want it to name the path", path, err.Error())
			}
		})
	}
}

// TestOpenRoster_ValidationRows proves B9's validation list, each row
// with its own written-out expectation.
func TestOpenRoster_ValidationRows(t *testing.T) {
	// atCapDisplay and overCapDisplay pad with "é" (a 2-byte, 1-rune
	// UTF-8 character) so a byte-length cap and a rune-count cap
	// disagree, the way internal/identity's own boundary fixture does:
	// 100 of them lands at 200 bytes (maxDisplayNameLength) but 100
	// runes, and one more crosses the byte cap while staying at 101
	// runes, far under a rune-based cap of 200.
	atCapDisplay := strings.Repeat("é", 100)
	overCapDisplay := strings.Repeat("é", 100) + "x"
	if len(atCapDisplay) != maxDisplayNameLength {
		t.Fatalf("fixture atCapDisplay is %d bytes, want exactly maxDisplayNameLength (%d)", len(atCapDisplay), maxDisplayNameLength)
	}
	if len(overCapDisplay) != maxDisplayNameLength+1 {
		t.Fatalf("fixture overCapDisplay is %d bytes, want exactly maxDisplayNameLength+1 (%d)", len(overCapDisplay), maxDisplayNameLength+1)
	}

	atCapUsername := strings.Repeat("a", maxStreamerUsernameLength)
	overCapUsername := strings.Repeat("a", maxStreamerUsernameLength+1)

	tests := []struct {
		name    string
		members string // the JSON "members" array body
		wantErr bool
		// wantSentinel is the specific refusal errors.Is must find
		// wrapped inside ErrRosterRecordInvalid; nil when wantErr is
		// false.
		wantSentinel error
	}{
		{
			name:         "a duplicate address differing only in case",
			members:      `[{"address":"Alice@Example.com","display_name":"A"},{"address":"alice@example.com","display_name":"B"}]`,
			wantErr:      true,
			wantSentinel: errDuplicateAddress,
		},
		{
			name:         "a duplicate streamer_username",
			members:      `[{"address":"a@example.com","display_name":"A","streamer_username":"same"},{"address":"b@example.com","display_name":"B","streamer_username":"same"}]`,
			wantErr:      true,
			wantSentinel: errDuplicateStreamerUsername,
		},
		{
			name:         "a control character in a display name",
			members:      `[{"address":"a@example.com","display_name":"A\u0007B"}]`,
			wantErr:      true,
			wantSentinel: errInvalidDisplayName,
		},
		{
			name:         "a C1 control character (CSI) in a display name",
			members:      `[{"address":"a@example.com","display_name":"A\u009bB"}]`,
			wantErr:      true,
			wantSentinel: errInvalidDisplayName,
		},
		{
			name:         "a no-break space in a display name",
			members:      `[{"address":"a@example.com","display_name":"A` + "\u00a0" + `B"}]`,
			wantErr:      true,
			wantSentinel: errInvalidDisplayName,
		},
		{
			name:         "a right-to-left override in a display name",
			members:      `[{"address":"a@example.com","display_name":"A` + "\u202e" + `B"}]`,
			wantErr:      true,
			wantSentinel: errInvalidDisplayName,
		},
		{
			name:         "a zero-width space in a display name",
			members:      `[{"address":"a@example.com","display_name":"A` + "\u200b" + `B"}]`,
			wantErr:      true,
			wantSentinel: errInvalidDisplayName,
		},
		{
			name:    "a zero-width non-joiner in a display name loads",
			members: `[{"address":"a@example.com","display_name":"A` + "\u200c" + `B"}]`,
			wantErr: false,
		},
		{
			name:    "a zero-width joiner in a display name loads",
			members: `[{"address":"a@example.com","display_name":"A` + "\u200d" + `B"}]`,
			wantErr: false,
		},
		{
			name:         "a control character in an address",
			members:      `[{"address":"a\u0007b@example.com","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:         "an unparseable address",
			members:      `[{"address":"not-an-address","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:         "a NEL control in a quoted address local part",
			members:      `[{"address":"\"a\u0085b\"@example.com","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:         "a C1 CSI control in a quoted address local part",
			members:      `[{"address":"\"a\u009bb\"@example.com","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:         "a right-to-left override in a quoted address local part",
			members:      `[{"address":"\"a` + "\u202e" + `b\"@example.com","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:         "a no-break space in a quoted address local part",
			members:      `[{"address":"\"a` + "\u00a0" + `b\"@example.com","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:         "a zero-width space in a quoted address local part",
			members:      `[{"address":"\"a` + "\u200b" + `b\"@example.com","display_name":"A"}]`,
			wantErr:      true,
			wantSentinel: errInvalidMemberAddress,
		},
		{
			name:    "display name exactly at the length cap",
			members: `[{"address":"a@example.com","display_name":"` + atCapDisplay + `"}]`,
			wantErr: false,
		},
		{
			name:         "display name one byte over the length cap",
			members:      `[{"address":"a@example.com","display_name":"` + overCapDisplay + `"}]`,
			wantErr:      true,
			wantSentinel: errInvalidDisplayName,
		},
		{
			name:    "streamer_username exactly at the length cap",
			members: `[{"address":"a@example.com","display_name":"A","streamer_username":"` + atCapUsername + `"}]`,
			wantErr: false,
		},
		{
			name:         "streamer_username one byte over the length cap",
			members:      `[{"address":"a@example.com","display_name":"A","streamer_username":"` + overCapUsername + `"}]`,
			wantErr:      true,
			wantSentinel: errInvalidStreamerUsername,
		},
		{
			name:         "a streamer_username carrying a non-ASCII-printable byte",
			members:      `[{"address":"a@example.com","display_name":"A","streamer_username":"na` + "é" + `me"}]`,
			wantErr:      true,
			wantSentinel: errInvalidStreamerUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRosterFile(t, `{"schema_version":1,"members":`+tt.members+`}`)
			_, err := OpenRoster(path)
			if tt.wantErr && err == nil {
				t.Fatalf("OpenRoster() error = nil, want a validation refusal")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("OpenRoster() error = %v, want nil", err)
			}
			if !tt.wantErr {
				return
			}
			if !errors.Is(err, ErrRosterRecordInvalid) {
				t.Errorf("OpenRoster() error = %v, want it to wrap ErrRosterRecordInvalid", err)
			}
			if !errors.Is(err, tt.wantSentinel) {
				t.Errorf("OpenRoster() error = %v, want it to wrap %v", err, tt.wantSentinel)
			}
		})
	}
}

// TestOpenRoster_NoFixtureAddressInError proves that no address or
// display name from a fixture ever appears in a load error: this
// package's doc comment promises no error carries a record's own
// value, with the roster audit record (Task 5's) the only exception.
func TestOpenRoster_NoFixtureAddressInError(t *testing.T) {
	const fixtureAddress = "duplicate-fixture@example.com"
	path := writeRosterFile(t, `{"schema_version":1,"members":[
		{"address":"`+fixtureAddress+`","display_name":"First"},
		{"address":"`+fixtureAddress+`","display_name":"Second"}
	]}`)

	_, err := OpenRoster(path)
	if err == nil {
		t.Fatal("OpenRoster() error = nil, want the duplicate-address refusal")
	}
	if strings.Contains(err.Error(), fixtureAddress) {
		t.Errorf("OpenRoster() error = %q, want it to not contain the fixture address", err.Error())
	}
}

// TestRoster_LookupErrorNamesNoAddress proves a Lookup refusal for an
// input net/mail cannot parse carries neither address in it: net/mail
// quotes the unparsed remainder of a two-address list, so wrapping its
// error would echo the second address back to the caller.
func TestRoster_LookupErrorNamesNoAddress(t *testing.T) {
	const first = "lookup-first@example.com"
	const second = "lookup-second@example.com"
	path := writeRosterFile(t, `{"schema_version":1,"members":[]}`)
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	_, err = roster.Lookup(first + ", " + second)
	if !errors.Is(err, ErrMemberNotFound) {
		t.Fatalf("Lookup() error = %v, want it to wrap ErrMemberNotFound", err)
	}
	for _, addr := range []string{first, second} {
		if strings.Contains(err.Error(), addr) {
			t.Errorf("Lookup() error = %q, want it to not contain %q", err.Error(), addr)
		}
	}
}

// TestNormalizeMemberAddress_SingleResolutionPoint proves R1's single
// resolution point by grep: normalizeMemberAddress is the only
// function in this file that resolves a roster member's address.
// validateActor calls identity.NormalizeAddress too, for a distinct
// purpose (shaping an audit actor, never a roster key), so the
// invariant is checked per function rather than as one file-wide
// count. Named mutation 5 (add a second member-address derivation
// outside normalizeMemberAddress) fails this test.
func TestNormalizeMemberAddress_SingleResolutionPoint(t *testing.T) {
	source, err := os.ReadFile("roster.go")
	if err != nil {
		t.Fatalf("read roster.go: %v", err)
	}
	callsPerFunc := regexp.MustCompile(`identity\.NormalizeAddress\(`)
	for _, fn := range []string{"normalizeMemberAddress", "validateActor"} {
		body := regexp.MustCompile(`(?s)func ` + fn + `\(.*?\n}`).Find(source)
		if body == nil {
			t.Fatalf("roster.go has no func %s", fn)
		}
		if n := len(callsPerFunc.FindAllIndex(body, -1)); n != 1 {
			t.Errorf("%s calls identity.NormalizeAddress %d times, want exactly 1", fn, n)
		}
	}
	if total := len(callsPerFunc.FindAllIndex(source, -1)); total != 2 {
		t.Errorf("roster.go calls identity.NormalizeAddress %d times, want exactly 2 (normalizeMemberAddress and validateActor)", total)
	}
}

// TestRoster_LookupCaseInsensitive proves that a lookup for an address
// differing only in case resolves to the same member (spec:1161-1163).
func TestRoster_LookupCaseInsensitive(t *testing.T) {
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"Alice@Example.com","display_name":"Alice"}]}`)
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	got, err := roster.Lookup("alice@EXAMPLE.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Address != "alice@example.com" {
		t.Errorf("Lookup().Address = %q, want %q", got.Address, "alice@example.com")
	}

	if _, err := roster.Lookup("nobody@example.com"); !errors.Is(err, ErrMemberNotFound) {
		t.Errorf("Lookup(unknown) error = %v, want ErrMemberNotFound", err)
	}
}

// TestRoster_TwoEntryPoints proves the read-open and the write-open
// refuse opposite things: the read-open refuses an absent file, the
// write-open creates one at 0600 and refuses an existing one. Named
// mutation 6 (let the write-open succeed against an existing file)
// fails the "refuses a second roster" case; named mutation 7 (let the
// read-open tolerate an absent file once the write-open exists) fails
// the "read-open still refuses after create" case.
func TestRoster_TwoEntryPoints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roster.json")

	if _, err := OpenRoster(path); !errors.Is(err, ErrRosterMissing) {
		t.Fatalf("OpenRoster(absent) error = %v, want ErrRosterMissing", err)
	}

	if _, err := CreateRoster(t.Context(), path); err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("CreateRoster() file mode = %o, want %o", mode, 0o600)
	}

	if _, err := CreateRoster(t.Context(), path); !errors.Is(err, ErrRosterExists) {
		t.Fatalf("CreateRoster(existing) error = %v, want ErrRosterExists", err)
	}

	loaded, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster(after create) error = %v, want nil: the read-open must still work once the file exists", err)
	}
	if got := mustList(t, loaded); len(got) != 0 {
		t.Errorf("List() after CreateRoster = %v, want empty", got)
	}
}

// TestRosterExampleFile proves config/roster.example.json parses
// through the loader, so the shipped example cannot rot.
func TestRosterExampleFile(t *testing.T) {
	roster, err := OpenRoster(filepath.Join("..", "..", "config", "roster.example.json"))
	if err != nil {
		t.Fatalf("OpenRoster(config/roster.example.json) error = %v, want nil", err)
	}
	if len(mustList(t, roster)) == 0 {
		t.Error("config/roster.example.json parsed to zero members, want at least one")
	}
}

// TestValidateWriteCandidate_StreamerUsernameShape proves the write
// path's narrower shape check: an alphanumeric first character, then
// any run of alphanumerics, '.', '_', '@', '+', or '-', within the
// loader's own ASCII-printable and length checks.
// Named mutation 5 (drop the alphanumeric first character) admits
// ".x" and "-rf"; named mutation 6 (drop '@' from the character
// class) refuses the email-shaped row.
func TestValidateWriteCandidate_StreamerUsernameShape(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{name: "letters, digits, and underscore", want: "a_b"},
		{name: "uppercase, dot, and hyphen", want: "A.b-c"},
		{name: "a single digit start", want: "x9"},
		{name: "email-shaped with a plus", want: "a+b@example.test"},
		{name: "email-shaped plain", want: "a@example.test"},
		{name: "leading dot", want: ".x", wantErr: true},
		{name: "only dots", want: "..", wantErr: true},
		{name: "leading hyphen", want: "-rf", wantErr: true},
		{name: "leading underscore", want: "_x", wantErr: true},
		{name: "a percent sign", want: "a%", wantErr: true},
		{name: "an embedded space", want: "a b", wantErr: true},
		{name: "a non-ASCII byte", want: "naéme", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := Member{Address: "a@example.com", DisplayName: "A", StreamerUsername: tt.want}
			_, err := validateWriteCandidate(candidate)
			if tt.wantErr {
				if !errors.Is(err, errInvalidStreamerUsername) {
					t.Errorf("validateWriteCandidate(%q) error = %v, want it to wrap errInvalidStreamerUsername", tt.want, err)
				}
				return
			}
			if err != nil {
				t.Errorf("validateWriteCandidate(%q) error = %v, want nil", tt.want, err)
			}
		})
	}
}

// TestValidateWriteCandidate_StreamerUsernameRequired proves the write
// path refuses a candidate with no streamer_username, while validate
// alone (the loader's path) still accepts one absent. Named mutation 7
// (move the requirement to the loader) fails the load half of this
// test.
func TestValidateWriteCandidate_StreamerUsernameRequired(t *testing.T) {
	candidate := Member{Address: "a@example.com", DisplayName: "A"}

	if _, err := validateWriteCandidate(candidate); !errors.Is(err, errMissingStreamerUsername) {
		t.Errorf("validateWriteCandidate(no streamer_username) error = %v, want it to wrap errMissingStreamerUsername", err)
	}
	if _, err := memberRecord(candidate).validate(); err != nil {
		t.Errorf("validate(no streamer_username) error = %v, want nil: the loader keeps it optional", err)
	}
}

// TestAddMember_StreamerUsernameCaseFoldDuplicate proves the write
// path's disk re-check folds case, so "Alice" beside an existing
// "alice" is refused rather than admitted as two principals mapping to
// one account. Named mutation 4 (compare with ==) admits this row.
func TestAddMember_StreamerUsernameCaseFoldDuplicate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"seed@example.com","display_name":"Seed","streamer_username":"alice"}]}`)
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	candidate := Member{Address: "new@example.com", DisplayName: "New", StreamerUsername: "Alice"}
	err = roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger)
	if !errors.Is(err, errDuplicateStreamerUsername) {
		t.Errorf("AddMember(%q beside existing %q) error = %v, want it to wrap errDuplicateStreamerUsername", "Alice", "alice", err)
	}
}

// TestOpenRoster_StreamerUsernameCaseFoldDuplicate proves the loader's
// uniqueness check also folds case, mirroring the write path.
func TestOpenRoster_StreamerUsernameCaseFoldDuplicate(t *testing.T) {
	path := writeRosterFile(t, `{"schema_version":1,"members":[
		{"address":"a@example.com","display_name":"A","streamer_username":"Alice"},
		{"address":"b@example.com","display_name":"B","streamer_username":"alice"}
	]}`)
	_, err := OpenRoster(path)
	if !errors.Is(err, errDuplicateStreamerUsername) {
		t.Errorf("OpenRoster(case-differing streamer_username) error = %v, want it to wrap errDuplicateStreamerUsername", err)
	}
}

// TestAddMember_ActorAndChannelRefusals proves AddMember refuses
// before any write when the channel is not one Channel.Valid accepts,
// or when the actor does not meet its channel's own rule: today the
// only reachable rule requires the actor to normalize as a login
// address, since every channel this package can currently assert is a
// network assertion. Named mutation 8 (skip the channel validity
// check) admits the Channel(0) row.
func TestAddMember_ActorAndChannelRefusals(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	candidate := Member{Address: "new@example.com", DisplayName: "New", StreamerUsername: "new"}

	tests := []struct {
		name    string
		actor   string
		channel identity.Channel
	}{
		{name: "an empty actor", actor: "", channel: identity.ChannelCookie},
		{name: "a malformed actor on the cookie channel", actor: "not-an-address", channel: identity.ChannelCookie},
		{name: "an invalid channel", actor: "admin@example.com", channel: identity.Channel(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRosterFile(t, `{"schema_version":1,"members":[]}`)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture before AddMember: %v", err)
			}
			roster, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster: %v", err)
			}

			err = roster.AddMember(t.Context(), candidate, tt.actor, tt.channel, "198.51.100.7", logger)
			if !errors.Is(err, errInvalidActor) {
				t.Errorf("AddMember() error = %v, want it to wrap errInvalidActor", err)
			}
			if !errors.Is(err, ErrRosterRecordInvalid) {
				t.Errorf("AddMember() error = %v, want it to also wrap ErrRosterRecordInvalid", err)
			}

			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture after AddMember: %v", err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("AddMember() on a refused actor or channel changed the file: before %q, after %q", before, after)
			}
		})
	}
}

// TestAddMember_ZeroWidthAddressVariantRefused proves that adding an
// address carrying a zero-width character is refused outright, so it
// can never land beside an existing member as a visually
// indistinguishable second entry. Fixture state: a roster already
// holding the plain address.
func TestAddMember_ZeroWidthAddressVariantRefused(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice","streamer_username":"alice"}]}`)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture before AddMember: %v", err)
	}
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	candidate := Member{Address: "\"a\u200dlice\"@example.com", DisplayName: "Alice", StreamerUsername: "alice-2"}
	err = roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger)
	if !errors.Is(err, ErrRosterRecordInvalid) {
		t.Errorf("AddMember(zero-width address variant) error = %v, want it to wrap ErrRosterRecordInvalid", err)
	}
	if !errors.Is(err, errInvalidMemberAddress) {
		t.Errorf("AddMember(zero-width address variant) error = %v, want it to wrap errInvalidMemberAddress", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after AddMember: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("AddMember(zero-width address variant) changed the file: before %q, after %q", before, after)
	}
}

// TestIsASCIIPrintable is a direct table-driven test of the helper
// TestOpenRoster_ValidationRows exercises indirectly.
func TestIsASCIIPrintable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "plain ASCII", input: "example-name", want: true},
		{name: "space is printable", input: "a name", want: true},
		{name: "control character", input: "a\tb", want: false},
		{name: "non-ASCII byte", input: "naéme", want: false},
		{name: "empty string", input: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isASCIIPrintable(tt.input); got != tt.want {
				t.Errorf("isASCIIPrintable(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestDisplayNameCarriesRefusedRune is a direct table-driven test of
// the helper: "Alice Example" is pinned to acceptance so the ordinary
// space never regresses to a refusal, and the zero-width joiner and
// non-joiner are pinned to acceptance so a later category-wide refusal
// cannot silently re-refuse them.
func TestDisplayNameCarriesRefusedRune(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "plain text", input: "Example Name", want: false},
		{name: "a name carrying an ordinary space", input: "Alice Example", want: false},
		{name: "bell character", input: "a\u0007b", want: true},
		{name: "DEL", input: "a\u007fb", want: true},
		{name: "newline", input: "a\nb", want: true},
		{name: "C1 CSI", input: "a\u009bb", want: true},
		{name: "non-control non-ASCII", input: "Zoë", want: false},
		{name: "no-break space (Zs)", input: "a\u00a0b", want: true},
		{name: "line separator (Zl)", input: "a\u2028b", want: true},
		{name: "paragraph separator (Zp)", input: "a\u2029b", want: true},
		{name: "right-to-left override (Cf)", input: "a\u202eb", want: true},
		{name: "zero-width space (Cf)", input: "a\u200bb", want: true},
		{name: "zero-width non-joiner (Cf, admitted)", input: "a\u200cb", want: false},
		{name: "zero-width joiner (Cf, admitted)", input: "a\u200db", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayNameCarriesRefusedRune(tt.input); got != tt.want {
				t.Errorf("displayNameCarriesRefusedRune(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestCreateRoster_PropagatesStatError proves CreateRoster reports a
// Stat failure other than absence, so a permission-denied parent
// directory is not silently treated as "file absent, create it".
func TestCreateRoster_PropagatesStatError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not block stat traversal")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	path := filepath.Join(dir, "roster.json")
	if _, err := CreateRoster(t.Context(), path); err == nil {
		t.Fatal("CreateRoster() error = nil, want a stat failure")
	} else if errors.Is(err, ErrRosterExists) {
		t.Errorf("CreateRoster() error = %v, want it to not claim the file exists", err)
	}
}

// TestAddMember_ByteIdentityOnFailure proves that a write refused for
// any reason leaves the file byte-identical: the validation failures
// caught before the lock, the collision re-check caught with the lock
// held, and the admin refusal are each a different code path, so each
// gets its own row. The admin refusal gets one row per channel, since
// no channel, a service token included, may set the flag. Named
// mutation 1 (write before validating) fails every row here; named
// mutation 6 (accept the admin flag on any assertion channel) fails
// each admin row.
func TestAddMember_ByteIdentityOnFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	const seedJSON = `{"schema_version":1,"members":[{"address":"seed@example.com","display_name":"Seed","streamer_username":"seed"}]}`

	tests := []struct {
		name      string
		candidate Member
		channel   identity.Channel
		wantErr   error
	}{
		{
			name:      "an address that does not parse",
			candidate: Member{Address: "not-an-address", DisplayName: "A"},
			channel:   identity.ChannelServiceToken,
			wantErr:   ErrRosterRecordInvalid,
		},
		{
			name:      "a duplicate address",
			candidate: Member{Address: "seed@example.com", DisplayName: "A", StreamerUsername: "not-seed"},
			channel:   identity.ChannelServiceToken,
			wantErr:   ErrRosterRecordInvalid,
		},
		{
			name:      "a duplicate streamer_username",
			candidate: Member{Address: "new@example.com", DisplayName: "A", StreamerUsername: "seed"},
			channel:   identity.ChannelServiceToken,
			wantErr:   ErrRosterRecordInvalid,
		},
		{
			name:      "the admin flag on a cookie channel",
			candidate: Member{Address: "new@example.com", DisplayName: "A", Admin: true},
			channel:   identity.ChannelCookie,
			wantErr:   ErrRosterAdminChannelForbidden,
		},
		{
			name:      "the admin flag on a header channel",
			candidate: Member{Address: "new@example.com", DisplayName: "A", Admin: true},
			channel:   identity.ChannelHeader,
			wantErr:   ErrRosterAdminChannelForbidden,
		},
		{
			name:      "the admin flag on a service-token channel",
			candidate: Member{Address: "new@example.com", DisplayName: "A", Admin: true},
			channel:   identity.ChannelServiceToken,
			wantErr:   ErrRosterAdminChannelForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRosterFile(t, seedJSON)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture before AddMember: %v", err)
			}
			roster, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster: %v", err)
			}

			err = roster.AddMember(t.Context(), tt.candidate, "tester@example.com", tt.channel, "203.0.113.1", logger)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddMember() error = %v, want it to wrap %v", err, tt.wantErr)
			}

			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture after AddMember: %v", err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("AddMember() on a failing candidate changed the file: before %q, after %q", before, after)
			}
		})
	}
}

// TestAddMember_WriteReReadAndMode proves the successful path: the
// write lands at 0600 through Task 4's write-open, a lookup right
// after AddMember answers from the re-read with no reopen, and the
// read-open still refuses an absent file elsewhere once the
// write-open's file exists. Named mutation 4 (skip the re-read) fails
// the lookup here; named mutation 7 (let the read-open tolerate an
// absent file) fails the OpenRoster(other) case.
func TestAddMember_WriteReReadAndMode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := filepath.Join(t.TempDir(), "roster.json")
	roster, err := CreateRoster(t.Context(), path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidate := Member{Address: "Alice@Example.com", DisplayName: "Alice", StreamerUsername: "alice"}
	if err := roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	got, err := roster.Lookup("alice@example.com")
	if err != nil {
		t.Fatalf("Lookup after AddMember: %v", err)
	}
	if got.DisplayName != "Alice" {
		t.Errorf("Lookup().DisplayName = %q, want %q", got.DisplayName, "Alice")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat after AddMember: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("file mode after AddMember = %o, want %o", mode, 0o600)
	}

	other := filepath.Join(filepath.Dir(path), "other-roster.json")
	if _, err := OpenRoster(other); !errors.Is(err, ErrRosterMissing) {
		t.Errorf("OpenRoster(%q) error = %v, want ErrRosterMissing", other, err)
	}
}

// TestAddMember_AuditRecord proves the audit record names every
// attribute the spec requires and carries no other key, so no token,
// secret, or cookie value can ride along. Named mutation 5 (drop the
// channel from the audit record) fails the "channel" key here.
func TestAddMember_AuditRecord(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	path := filepath.Join(t.TempDir(), "roster.json")
	roster, err := CreateRoster(t.Context(), path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidate := Member{Address: "alice@example.com", DisplayName: "Alice", StreamerUsername: "alice"}
	if err := roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logBuf.Bytes()), &record); err != nil {
		t.Fatalf("unmarshal audit record %q: %v", logBuf.String(), err)
	}

	want := map[string]any{
		"actor":     "admin@example.com",
		"channel":   "cookie",
		"address":   "alice@example.com",
		"admin":     false,
		"client_ip": "198.51.100.7",
	}
	for key, wantValue := range want {
		gotValue, ok := record[key]
		if !ok {
			t.Errorf("audit record missing %q", key)
			continue
		}
		if gotValue != wantValue {
			t.Errorf("audit record[%q] = %v, want %v", key, gotValue, wantValue)
		}
	}

	allowed := map[string]bool{
		"time": true, "level": true, "msg": true,
		"actor": true, "channel": true, "address": true, "admin": true, "client_ip": true,
	}
	for key, value := range record {
		if !allowed[key] {
			t.Errorf("audit record carries unexpected key %q (value %v)", key, value)
		}
	}
}

// TestAddMember_ConcurrentDistinctMembersBothSurvive proves that two
// concurrent adds for distinct members both succeed and the file
// carries both, under -race. Named mutation 2 (take the lock after
// the re-check) does not reach this test, since these two candidates
// never collide; TestAddMember_ConcurrentStreamerUsernameCollision
// covers that.
func TestAddMember_ConcurrentDistinctMembersBothSurvive(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := filepath.Join(t.TempDir(), "roster.json")
	roster, err := CreateRoster(t.Context(), path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice", StreamerUsername: "alice"},
		{Address: "bob@example.com", DisplayName: "Bob", StreamerUsername: "bob"},
	}
	errs := make([]error, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate Member) {
			defer wg.Done()
			errs[i] = roster.AddMember(t.Context(), candidate, "tester@example.com", identity.ChannelServiceToken, "203.0.113.1", logger)
		}(i, candidate)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("AddMember(%d) error = %v, want nil", i, err)
		}
	}

	if got := mustList(t, roster); len(got) != len(candidates) {
		t.Fatalf("List() after concurrent adds = %d members, want %d", len(got), len(candidates))
	}
	onDisk, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster after concurrent adds: %v", err)
	}
	if got := mustList(t, onDisk); len(got) != len(candidates) {
		t.Errorf("the file on disk carries %d members, want %d", len(got), len(candidates))
	}
}

// TestAddMember_ConcurrentStreamerUsernameCollision proves that two
// concurrent adds claiming the same streamer_username leave exactly
// one member, with the other refused by ErrRosterRecordInvalid. Named
// mutation 2 (take the lock after the re-check) admits both, so this
// test fails under that mutation.
func TestAddMember_ConcurrentStreamerUsernameCollision(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := filepath.Join(t.TempDir(), "roster.json")
	roster, err := CreateRoster(t.Context(), path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice", StreamerUsername: "same"},
		{Address: "bob@example.com", DisplayName: "Bob", StreamerUsername: "same"},
	}
	errs := make([]error, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate Member) {
			defer wg.Done()
			errs[i] = roster.AddMember(t.Context(), candidate, "tester@example.com", identity.ChannelServiceToken, "203.0.113.1", logger)
		}(i, candidate)
	}
	wg.Wait()

	successes, refusals := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrRosterRecordInvalid):
			refusals++
		default:
			t.Errorf("AddMember() error = %v, want nil or ErrRosterRecordInvalid", err)
		}
	}
	if successes != 1 || refusals != 1 {
		t.Fatalf("got %d successes and %d refusals, want exactly one of each", successes, refusals)
	}
	if got := mustList(t, roster); len(got) != 1 {
		t.Fatalf("List() after the collision = %d members, want 1", len(got))
	}
}

// TestOpenRoster_StrictLoader proves the loader refuses every loose
// shape of the file rather than reading it as something it is not. A
// case-variant key ("ADMIN") and a repeated key would otherwise bind to
// the admin flag the way a lenient decoder binds them, and a null or
// missing members array would otherwise load as an empty roster. Each
// refusal is ErrRosterUnparseable naming the path; the version and
// members rows also name the rule that failed.
func TestOpenRoster_StrictLoader(t *testing.T) {
	const member = `{"address":"a@example.com","display_name":"A"}`
	tests := []struct {
		name        string
		content     string
		wantErr     error
		wantMention string
	}{
		{name: "zero-byte file", content: ``, wantErr: ErrRosterUnparseable},
		{name: "null", content: `null`, wantErr: ErrRosterUnparseable, wantMention: "schema_version"},
		{name: "empty object", content: `{}`, wantErr: ErrRosterUnparseable, wantMention: "schema_version"},
		{name: "version with no members", content: `{"schema_version":1}`, wantErr: ErrRosterUnparseable, wantMention: "members"},
		{name: "null members", content: `{"schema_version":1,"members":null}`, wantErr: ErrRosterUnparseable, wantMention: "members"},
		{name: "schema_version 0", content: `{"schema_version":0,"members":[]}`, wantErr: ErrRosterUnparseable, wantMention: "schema_version"},
		{name: "schema_version 2", content: `{"schema_version":2,"members":[]}`, wantErr: ErrRosterUnparseable, wantMention: "schema_version"},
		{name: "schema_version 99", content: `{"schema_version":99,"members":[]}`, wantErr: ErrRosterUnparseable, wantMention: "schema_version"},
		{
			name:    "an unknown member key",
			content: `{"schema_version":1,"members":[{"address":"a@example.com","display_name":"A","role":"owner"}]}`,
			wantErr: ErrRosterUnparseable,
		},
		{
			name:    "a case-variant admin key",
			content: `{"schema_version":1,"members":[{"address":"a@example.com","display_name":"A","ADMIN":true}]}`,
			wantErr: ErrRosterUnparseable,
		},
		{
			name:    "a duplicate admin key",
			content: `{"schema_version":1,"members":[{"address":"a@example.com","display_name":"A","admin":false,"admin":true}]}`,
			wantErr: ErrRosterUnparseable,
		},
		{
			name:    "an unknown top-level key",
			content: `{"schema_version":1,"members":[` + member + `],"admins":[]}`,
			wantErr: ErrRosterUnparseable,
		},
		{
			name:    "invalid UTF-8 in a display name",
			content: `{"schema_version":1,"members":[{"address":"a@example.com","display_name":"A` + "\xff" + `"}]}`,
			wantErr: ErrRosterUnparseable,
		},
		{
			name:    "trailing data after the object",
			content: `{"schema_version":1,"members":[` + member + `]} {}`,
			wantErr: ErrRosterUnparseable,
		},
		{name: "an empty members array loads", content: `{"schema_version":1,"members":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRosterFile(t, tt.content)
			roster, err := OpenRoster(path)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("OpenRoster() error = %v, want nil", err)
				}
				if got := mustList(t, roster); len(got) != 0 {
					t.Errorf("List() = %v, want empty", got)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("OpenRoster() error = %v, want it to wrap %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), path) {
				t.Errorf("OpenRoster() error = %q, want it to name the path", err.Error())
			}
			if tt.wantMention != "" && !strings.Contains(err.Error(), tt.wantMention) {
				t.Errorf("OpenRoster() error = %q, want it to name the %s rule", err.Error(), tt.wantMention)
			}
		})
	}
}

// replaceRosterFile writes content to a sibling of path and renames it
// over path, the way an atomic writer or an editor that saves by rename
// does, so path names a new inode.
func replaceRosterFile(t *testing.T, path, content string) {
	t.Helper()
	tmp := path + ".new"
	if err := os.WriteFile(tmp, []byte(content), 0o600); err != nil {
		t.Fatalf("write replacement: %v", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatalf("rename replacement: %v", err)
	}
}

// TestRoster_ReloadsChangedFile changes the file under an open Roster
// and asks it again. A roster that answered from the copy it loaded
// would keep resolving a member an operator removed by hand, so every
// change must reach Lookup, and a change that breaks the file must
// refuse rather than fall back to the last good copy.
func TestRoster_ReloadsChangedFile(t *testing.T) {
	const aliceAndBob = `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice"},{"address":"bob@example.com","display_name":"Bob"}]}`
	const bobOnly = `{"schema_version":1,"members":[{"address":"bob@example.com","display_name":"Bob"}]}`

	tests := []struct {
		name      string
		change    func(t *testing.T, path string)
		wantAlice error // nil means alice resolves
		wantBob   error
	}{
		{
			name:      "a hand removal rewritten in place",
			change:    func(t *testing.T, path string) { writeInPlace(t, path, bobOnly) },
			wantAlice: ErrMemberNotFound,
		},
		{
			name:      "a hand removal saved by rename",
			change:    func(t *testing.T, path string) { replaceRosterFile(t, path, bobOnly) },
			wantAlice: ErrMemberNotFound,
		},
		{
			name:      "an unparseable edit",
			change:    func(t *testing.T, path string) { writeInPlace(t, path, `{"schema_version":1,"members":[`) },
			wantAlice: ErrRosterUnparseable,
			wantBob:   ErrRosterUnparseable,
		},
		{
			name: "an edit failing validation",
			change: func(t *testing.T, path string) {
				writeInPlace(t, path, `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice"},{"address":"ALICE@example.com","display_name":"Again"}]}`)
			},
			wantAlice: ErrRosterRecordInvalid,
			wantBob:   ErrRosterRecordInvalid,
		},
		{
			name: "the file removed",
			change: func(t *testing.T, path string) {
				if err := os.Remove(path); err != nil {
					t.Fatalf("Remove: %v", err)
				}
			},
			wantAlice: ErrRosterMissing,
			wantBob:   ErrRosterMissing,
		},
		{
			name:   "the same bytes saved by rename",
			change: func(t *testing.T, path string) { replaceRosterFile(t, path, aliceAndBob) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRosterFile(t, aliceAndBob)
			roster, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster: %v", err)
			}
			if _, err := roster.Lookup("alice@example.com"); err != nil {
				t.Fatalf("Lookup(alice) before the change: %v", err)
			}

			tt.change(t, path)

			for _, row := range []struct {
				address string
				want    error
			}{{"alice@example.com", tt.wantAlice}, {"bob@example.com", tt.wantBob}} {
				_, err := roster.Lookup(row.address)
				if row.want == nil && err != nil {
					t.Errorf("Lookup(%s) error = %v, want nil", row.address, err)
				}
				if row.want != nil && !errors.Is(err, row.want) {
					t.Errorf("Lookup(%s) error = %v, want it to wrap %v", row.address, err, row.want)
				}
			}
			if tt.wantBob != nil {
				if _, err := roster.List(); !errors.Is(err, tt.wantBob) {
					t.Errorf("List() error = %v, want it to wrap %v", err, tt.wantBob)
				}
			}
		})
	}
}

// writeInPlace rewrites path's content through the same inode, the way
// an editor that saves in place does.
func writeInPlace(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("rewrite in place: %v", err)
	}
}

// TestRoster_ModTimeTrustWindow rewrites the file in place with the same
// inode, the same size, and the same modification time, so its stat
// cannot tell the old content from the new. A roster that read the file
// long after its modification time serves its cache, which is the only
// thing that explains the old answer surviving. A roster that read the
// file within a second of that time cannot trust the stat and re-reads,
// which is the only thing that explains the new answer.
func TestRoster_ModTimeTrustWindow(t *testing.T) {
	// alice and carol are the same length, so the rewrite keeps the size.
	const before = `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"A"}]}`
	const after = `{"schema_version":1,"members":[{"address":"carol@example.com","display_name":"A"}]}`
	if len(before) != len(after) {
		t.Fatalf("fixtures differ in length: %d and %d", len(before), len(after))
	}

	tests := []struct {
		name    string
		age     bool
		wantOld bool
	}{
		{name: "a read well after the modification time serves its cache", age: true, wantOld: true},
		{name: "a read within a second of the modification time re-reads", age: false, wantOld: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRosterFile(t, before)
			if tt.age {
				past := time.Now().Add(-10 * time.Second)
				if err := os.Chtimes(path, past, past); err != nil {
					t.Fatalf("Chtimes: %v", err)
				}
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Stat: %v", err)
			}
			roster, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster: %v", err)
			}

			writeInPlace(t, path, after)
			if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
				t.Fatalf("Chtimes restore: %v", err)
			}
			rewritten, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Stat after rewrite: %v", err)
			}
			if !os.SameFile(info, rewritten) || info.Size() != rewritten.Size() || !info.ModTime().Equal(rewritten.ModTime()) {
				t.Fatal("fixture: the rewrite changed the inode, the size, or the modification time")
			}

			_, aliceErr := roster.Lookup("alice@example.com")
			_, carolErr := roster.Lookup("carol@example.com")
			if tt.wantOld {
				if aliceErr != nil || !errors.Is(carolErr, ErrMemberNotFound) {
					t.Errorf("Lookup(alice) = %v, Lookup(carol) = %v; want the cached answer (alice resolves, carol does not)", aliceErr, carolErr)
				}
				return
			}
			if carolErr != nil || !errors.Is(aliceErr, ErrMemberNotFound) {
				t.Errorf("Lookup(alice) = %v, Lookup(carol) = %v; want the re-read answer (carol resolves, alice does not)", aliceErr, carolErr)
			}
		})
	}
}

// TestRoster_CrossProcessWrites runs the command line and the server as
// two Roster values over one file, both opened before either writes, so
// the server's copy in memory is stale. The command line adds a member,
// an operator removes one by hand, and then the server adds a member.
// The final file must carry both additions and not the removed member:
// a write merged from memory resurrects the removal and drops the
// command line's addition. One row asks the server about the removed
// member before its own add, which only a re-read can answer; the other
// adds with the stale copy untouched.
func TestRoster_CrossProcessWrites(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	const seed = `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice"},{"address":"bob@example.com","display_name":"Bob"}]}`
	const bobRemoved = `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice"},{"address":"carol@example.com","display_name":"Carol"}]}`

	for _, lookupFirst := range []bool{true, false} {
		name := "the server adds with its stale copy"
		if lookupFirst {
			name = "the server looks up the removed member before adding"
		}
		t.Run(name, func(t *testing.T) {
			path := writeRosterFile(t, seed)
			cli, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster(cli): %v", err)
			}
			server, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster(server): %v", err)
			}

			if err := cli.AddMember(t.Context(), Member{Address: "carol@example.com", DisplayName: "Carol", StreamerUsername: "carol"}, "operator@example.com", identity.ChannelServiceToken, "", logger); err != nil {
				t.Fatalf("cli AddMember: %v", err)
			}
			writeInPlace(t, path, bobRemoved)

			if lookupFirst {
				if _, err := server.Lookup("bob@example.com"); !errors.Is(err, ErrMemberNotFound) {
					t.Errorf("server Lookup(bob) after the hand removal error = %v, want ErrMemberNotFound", err)
				}
			}
			if err := server.AddMember(t.Context(), Member{Address: "dave@example.com", DisplayName: "Dave", StreamerUsername: "dave"}, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
				t.Fatalf("server AddMember: %v", err)
			}

			onDisk, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster(final): %v", err)
			}
			var got []string
			for _, m := range mustList(t, onDisk) {
				got = append(got, m.Address)
			}
			want := []string{"alice@example.com", "carol@example.com", "dave@example.com"}
			if !slices.Equal(got, want) {
				t.Errorf("final file members = %v, want %v", got, want)
			}
			if _, err := server.Lookup("bob@example.com"); !errors.Is(err, ErrMemberNotFound) {
				t.Errorf("server Lookup(bob) after its add error = %v, want ErrMemberNotFound", err)
			}
		})
	}
}

// TestRoster_ConcurrentWritersAcrossValues runs one writer per Roster
// value, all over one file and all at once, the way several processes
// would. Every addition must survive: the writers exclude one another
// through the directory lock alone, since no two share a mutex.
func TestRoster_ConcurrentWritersAcrossValues(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := writeRosterFile(t, `{"schema_version":1,"members":[]}`)

	const writers = 8
	rosters := make([]*Roster, writers)
	for i := range rosters {
		r, err := OpenRoster(path)
		if err != nil {
			t.Fatalf("OpenRoster(%d): %v", i, err)
		}
		rosters[i] = r
	}

	start := make(chan struct{})
	errs := make([]error, writers)
	var wg sync.WaitGroup
	for i, r := range rosters {
		wg.Go(func() {
			<-start
			candidate := Member{Address: fmt.Sprintf("writer%d@example.com", i), DisplayName: "Writer", StreamerUsername: fmt.Sprintf("writer%d", i)}
			errs[i] = r.AddMember(t.Context(), candidate, "operator@example.com", identity.ChannelServiceToken, "", logger)
		})
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("AddMember(writer %d) error = %v, want nil", i, err)
		}
	}
	onDisk, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster(final): %v", err)
	}
	if got := mustList(t, onDisk); len(got) != writers {
		t.Errorf("the file carries %d members, want %d: a concurrent write was lost", len(got), writers)
	}
}

// TestCreateRoster_ConcurrentCreate races two CreateRoster calls at one
// absent path, many times over. Exactly one must create the file and the
// other must get ErrRosterExists, and the file on disk must be the one
// the winner read back. A check made outside the lock lets both see the
// path absent, and the second write then replaces the first.
func TestCreateRoster_ConcurrentCreate(t *testing.T) {
	const rounds = 50
	for round := range rounds {
		path := filepath.Join(t.TempDir(), "roster.json")
		start := make(chan struct{})
		var (
			wg      sync.WaitGroup
			rosters [2]*Roster
			errs    [2]error
		)
		for i := range 2 {
			wg.Go(func() {
				<-start
				rosters[i], errs[i] = CreateRoster(t.Context(), path)
			})
		}
		close(start)
		wg.Wait()

		var winner *Roster
		exists := 0
		for i, err := range errs {
			switch {
			case err == nil:
				winner = rosters[i]
			case errors.Is(err, ErrRosterExists):
				exists++
			default:
				t.Fatalf("round %d: CreateRoster() error = %v, want nil or ErrRosterExists", round, err)
			}
		}
		if winner == nil || exists != 1 {
			t.Fatalf("round %d: %d ErrRosterExists and winner %v, want exactly one creator and one ErrRosterExists", round, exists, winner != nil)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("round %d: Stat: %v", round, err)
		}
		if !os.SameFile(info, winner.snap.info) {
			t.Fatalf("round %d: the file on disk is not the one the winner created", round)
		}
	}
}

// waitObservedContext closes waiting the first time Done is called. A
// roster writer consults Done only after an attempt on the directory
// lock has failed, so waiting closing proves the writer is inside its
// lock wait.
type waitObservedContext struct {
	context.Context
	once    sync.Once
	waiting chan struct{}
}

func (c *waitObservedContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

// TestAddMember_BusyWhileLockHeld holds the roster directory's lock
// from outside the Roster, as another process would, and starts a
// write. The write must give up with the busy error when its context
// ends and leave the file byte-identical. While it waits, a Lookup on
// the same Roster value must answer at once: the writer takes the
// directory lock before the in-process mutex, so a login never queues
// behind another process's write.
func TestAddMember_BusyWhileLockHeld(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice"}]}`)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		t.Fatalf("open roster directory: %v", err)
	}
	defer func() { _ = dir.Close() }()
	if err := syscall.Flock(int(dir.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock roster directory: %v", err)
	}

	timeout, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	ctx := &waitObservedContext{Context: timeout, waiting: make(chan struct{})}
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- roster.AddMember(ctx, Member{Address: "bob@example.com", DisplayName: "Bob", StreamerUsername: "bob"}, "operator@example.com", identity.ChannelServiceToken, "", logger)
	}()
	select {
	case <-ctx.waiting:
	case err := <-writeDone:
		t.Fatalf("AddMember() returned %v without waiting on the held directory lock", err)
	}

	if _, err := roster.Lookup("alice@example.com"); err != nil {
		t.Errorf("Lookup(alice) while a writer waits error = %v, want nil", err)
	}
	if timeout.Err() != nil {
		t.Error("Lookup returned only after the writer's context ended: it queued behind the writer's lock wait")
	}

	if err := <-writeDone; !errors.Is(err, errRosterBusy) {
		t.Errorf("AddMember() under a held lock error = %v, want it to wrap errRosterBusy", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("a refused write changed the file: before %q, after %q", before, after)
	}
}
