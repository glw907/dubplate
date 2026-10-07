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

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/engine/store"
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

// bootstrapAdmin is the first member newAdminRoster creates, and the
// admin every cookie-channel write in these tests names as its actor.
var bootstrapAdmin = Member{Address: "admin@example.com", DisplayName: "Admin", Username: "admin", Admin: true}

// newAdminRoster creates a roster under t.TempDir() through CreateRoster
// carrying bootstrapAdmin alone, and returns it with its path.
func newAdminRoster(t *testing.T) (*Roster, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "roster.json")
	roster, err := CreateRoster(t.Context(), path, bootstrapAdmin, "operator", slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}
	return roster, path
}

// mustList returns r's members, failing the test when the roster does
// not load.
func mustList(t *testing.T, r *Roster) []Member {
	t.Helper()
	members, err := r.List(t.Context())
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

	atCapUsername := strings.Repeat("a", maxUsernameLength)
	overCapUsername := strings.Repeat("a", maxUsernameLength+1)

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
			name:         "a duplicate username",
			members:      `[{"address":"a@example.com","display_name":"A","username":"same"},{"address":"b@example.com","display_name":"B","username":"same"}]`,
			wantErr:      true,
			wantSentinel: errDuplicateUsername,
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
			name:    "username exactly at the length cap",
			members: `[{"address":"a@example.com","display_name":"A","username":"` + atCapUsername + `"}]`,
			wantErr: false,
		},
		{
			name:         "username one byte over the length cap",
			members:      `[{"address":"a@example.com","display_name":"A","username":"` + overCapUsername + `"}]`,
			wantErr:      true,
			wantSentinel: errInvalidUsername,
		},
		{
			name:         "a username carrying a non-ASCII-printable byte",
			members:      `[{"address":"a@example.com","display_name":"A","username":"na` + "é" + `me"}]`,
			wantErr:      true,
			wantSentinel: errInvalidUsername,
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

	_, err = roster.Lookup(t.Context(), first+", "+second)
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

	got, err := roster.Lookup(t.Context(), "alice@EXAMPLE.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Address != "alice@example.com" {
		t.Errorf("Lookup().Address = %q, want %q", got.Address, "alice@example.com")
	}

	if _, err := roster.Lookup(t.Context(), "nobody@example.com"); !errors.Is(err, ErrMemberNotFound) {
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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := CreateRoster(t.Context(), path, bootstrapAdmin, "operator", logger); err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("CreateRoster() file mode = %o, want %o", mode, 0o600)
	}

	second := Member{Address: "other@example.com", DisplayName: "Other", Username: "other", Admin: true}
	if _, err := CreateRoster(t.Context(), path, second, "operator", logger); !errors.Is(err, ErrRosterExists) {
		t.Fatalf("CreateRoster(existing) error = %v, want ErrRosterExists", err)
	}

	loaded, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster(after create) error = %v, want nil: the read-open must still work once the file exists", err)
	}
	if got := mustList(t, loaded); !slices.Equal(got, []Member{bootstrapAdmin}) {
		t.Errorf("List() after CreateRoster = %v, want only the first member", got)
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

// TestValidateWriteCandidate_UsernameShape proves the write
// path's narrower shape check: an alphanumeric first character, then
// any run of alphanumerics, '.', '_', '@', '+', or '-', within the
// loader's own ASCII-printable and length checks.
// Named mutation 5 (drop the alphanumeric first character) admits
// ".x" and "-rf"; named mutation 6 (drop '@' from the character
// class) refuses the email-shaped row.
func TestValidateWriteCandidate_UsernameShape(t *testing.T) {
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
			candidate := Member{Address: "a@example.com", DisplayName: "A", Username: tt.want}
			_, err := validateWriteCandidate(candidate)
			if tt.wantErr {
				if !errors.Is(err, errInvalidUsername) {
					t.Errorf("validateWriteCandidate(%q) error = %v, want it to wrap errInvalidUsername", tt.want, err)
				}
				return
			}
			if err != nil {
				t.Errorf("validateWriteCandidate(%q) error = %v, want nil", tt.want, err)
			}
		})
	}
}

// TestValidateWriteCandidate_UsernameRequired proves the write
// path refuses a candidate with no username, while validate
// alone (the loader's path) still accepts one absent. Named mutation 7
// (move the requirement to the loader) fails the load half of this
// test.
func TestValidateWriteCandidate_UsernameRequired(t *testing.T) {
	candidate := Member{Address: "a@example.com", DisplayName: "A"}

	if _, err := validateWriteCandidate(candidate); !errors.Is(err, errMissingUsername) {
		t.Errorf("validateWriteCandidate(no username) error = %v, want it to wrap errMissingUsername", err)
	}
	if _, err := recordOf(candidate).validate(); err != nil {
		t.Errorf("validate(no username) error = %v, want nil: the loader keeps it optional", err)
	}
}

// TestAddMember_UsernameCaseFoldDuplicate proves the write
// path's disk re-check folds case, so "Alice" beside an existing
// "alice" is refused rather than admitted as two principals mapping to
// one account. Named mutation 4 (compare with ==) admits this row.
func TestAddMember_UsernameCaseFoldDuplicate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"admin@example.com","display_name":"Admin","username":"alice","admin":true}]}`)
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	candidate := Member{Address: "new@example.com", DisplayName: "New", Username: "Alice"}
	err = roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger)
	if !errors.Is(err, errDuplicateUsername) {
		t.Errorf("AddMember(%q beside existing %q) error = %v, want it to wrap errDuplicateUsername", "Alice", "alice", err)
	}
}

// TestOpenRoster_UsernameCaseFoldDuplicate proves the loader's
// uniqueness check also folds case, mirroring the write path.
func TestOpenRoster_UsernameCaseFoldDuplicate(t *testing.T) {
	path := writeRosterFile(t, `{"schema_version":1,"members":[
		{"address":"a@example.com","display_name":"A","username":"Alice"},
		{"address":"b@example.com","display_name":"B","username":"alice"}
	]}`)
	_, err := OpenRoster(path)
	if !errors.Is(err, errDuplicateUsername) {
		t.Errorf("OpenRoster(case-differing username) error = %v, want it to wrap errDuplicateUsername", err)
	}
}

// TestAddMember_ActorAndChannelRefusals proves AddMember refuses
// before any write when the channel is not one Channel.Valid accepts,
// or when the actor does not meet its channel's own rule: a cookie
// actor must normalize as a login address, and a command-line actor
// must be non-empty and carry no rune an address refuses. Named
// mutation 8 (skip the channel validity check) admits the Channel(0)
// row.
func TestAddMember_ActorAndChannelRefusals(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	candidate := Member{Address: "new@example.com", DisplayName: "New", Username: "new"}

	tests := []struct {
		name    string
		actor   string
		channel identity.Channel
	}{
		{name: "an empty actor", actor: "", channel: identity.ChannelCookie},
		{name: "a malformed actor on the cookie channel", actor: "not-an-address", channel: identity.ChannelCookie},
		{name: "an invalid channel", actor: "admin@example.com", channel: identity.Channel(0)},
		{name: "an empty actor on the command line", actor: "", channel: identity.ChannelCommandLine},
		{name: "a command-line actor carrying a space", actor: "some operator", channel: identity.ChannelCommandLine},
		{name: "a command-line actor carrying U+202E", actor: "oper\u202eator", channel: identity.ChannelCommandLine},
		{name: "a command-line actor carrying a newline", actor: "operator\nforged", channel: identity.ChannelCommandLine},
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
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"alice@example.com","display_name":"Alice","username":"alice"}]}`)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture before AddMember: %v", err)
	}
	roster, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	candidate := Member{Address: "\"a\u200dlice\"@example.com", DisplayName: "Alice", Username: "alice-2"}
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

// defaultIgnorableRoster is the same ten code points
// internal/identity's TestNormalizeAddress_DefaultIgnorableRefused
// pins, checked here against the roster's own two write paths: the
// loader and AddMember on both channels. The three preExisting rows
// were already refused, before the derived-set refusal existed, by the
// loader's and AddMember's shared reliance on identity.NormalizeAddress,
// which already refused the whole Cf category; they are labelled
// regressions rather than proofs of the derived set's own reach.
var defaultIgnorableRoster = []struct {
	name        string
	r           rune
	preExisting bool
}{
	{name: "Hangul filler U+115F", r: 'ᅟ'},
	{name: "Hangul jungseong filler U+1160", r: 'ᅠ'},
	{name: "Hangul filler U+3164", r: 'ㅤ'},
	{name: "halfwidth Hangul filler U+FFA0", r: 'ﾠ'},
	{name: "combining grapheme joiner U+034F", r: '͏'},
	{name: "variation selector 16 U+FE0F", r: '️'},
	{name: "zero width space U+200B", r: '​', preExisting: true},
	{name: "soft hyphen U+00AD", r: '­', preExisting: true},
	{name: "tag digit one U+E0001", r: '\U000E0001', preExisting: true},
	{name: "braille pattern blank U+2800", r: '⠀'},
}

// defaultIgnorableAddress returns a quoted-local-part address carrying
// r between two letters, the same shape TestOpenRoster_ValidationRows
// uses for its own invisible-rune rows.
func defaultIgnorableAddress(r rune) string {
	return "\"a" + string(r) + "b\"@example.com"
}

// TestOpenRoster_DefaultIgnorableAddressRefused proves the derived-set
// refusal at the roster loader: each of the ten code points, embedded
// in an address, refuses the whole roster at load. Fixture state: the
// same address with no code point loads a one-member roster, so the
// refusal below is the code point's and not the quoting.
func TestOpenRoster_DefaultIgnorableAddressRefused(t *testing.T) {
	plainPath := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"ab@example.com","display_name":"A"}]}`)
	if _, err := OpenRoster(plainPath); err != nil {
		t.Fatalf("OpenRoster(plain fixture) = %v, want nil", err)
	}

	for _, tt := range defaultIgnorableRoster {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(struct {
				SchemaVersion int            `json:"schema_version"`
				Members       []memberRecord `json:"members"`
			}{SchemaVersion: 1, Members: []memberRecord{{Address: defaultIgnorableAddress(tt.r), DisplayName: "A"}}})
			if err != nil {
				t.Fatalf("json.Marshal fixture: %v", err)
			}
			path := writeRosterFile(t, string(body))
			_, err = OpenRoster(path)
			if err == nil {
				t.Fatalf("OpenRoster() error = nil, want a refusal")
			}
			if !errors.Is(err, ErrRosterRecordInvalid) {
				t.Errorf("OpenRoster() error = %v, want it to wrap ErrRosterRecordInvalid", err)
			}
			if !errors.Is(err, errInvalidMemberAddress) {
				t.Errorf("OpenRoster() error = %v, want it to wrap errInvalidMemberAddress", err)
			}
		})
	}
}

// TestAddMember_DefaultIgnorableAddressRefused proves the derived-set
// refusal at AddMember, on both channels: a new member's address
// carrying one of the ten code points is refused and changes no byte on
// disk. Fixture state: the same address with no code point is accepted
// by AddMember on both channels, proved once each, so the refusal below
// is the code point's.
func TestAddMember_DefaultIgnorableAddressRefused(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range []struct {
		channel  identity.Channel
		actor    string
		address  string
		streamer string
	}{
		{identity.ChannelCookie, "admin@example.com", "ab@example.com", "a"},
		{identity.ChannelCommandLine, "operator", "cd@example.com", "c"},
	} {
		plainPath := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"admin@example.com","display_name":"Admin","username":"admin","admin":true}]}`)
		plainRoster, err := OpenRoster(plainPath)
		if err != nil {
			t.Fatalf("OpenRoster: %v", err)
		}
		plainCandidate := Member{Address: tt.address, DisplayName: "A", Username: tt.streamer}
		if err := plainRoster.AddMember(t.Context(), plainCandidate, tt.actor, tt.channel, "198.51.100.7", logger); err != nil {
			t.Fatalf("AddMember(plain fixture, %s) = %v, want nil", tt.channel, err)
		}
	}

	for _, tt := range defaultIgnorableRoster {
		for _, channel := range []identity.Channel{identity.ChannelCookie, identity.ChannelCommandLine} {
			t.Run(tt.name+"/"+channel.String(), func(t *testing.T) {
				path := writeRosterFile(t, `{"schema_version":1,"members":[]}`)
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read fixture before AddMember: %v", err)
				}
				roster, err := OpenRoster(path)
				if err != nil {
					t.Fatalf("OpenRoster: %v", err)
				}

				candidate := Member{Address: defaultIgnorableAddress(tt.r), DisplayName: "A", Username: "a"}
				var actor string
				if channel == identity.ChannelCommandLine {
					actor = "operator"
				} else {
					actor = "admin@example.com"
				}
				err = roster.AddMember(t.Context(), candidate, actor, channel, "198.51.100.7", logger)
				if err == nil {
					t.Fatalf("AddMember() error = nil, want a refusal")
				}
				if !errors.Is(err, ErrRosterRecordInvalid) {
					t.Errorf("AddMember() error = %v, want it to wrap ErrRosterRecordInvalid", err)
				}
				if !errors.Is(err, errInvalidMemberAddress) {
					t.Errorf("AddMember() error = %v, want it to wrap errInvalidMemberAddress", err)
				}

				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read fixture after AddMember: %v", err)
				}
				if !bytes.Equal(before, after) {
					t.Errorf("AddMember() on a refused address changed the file: before %q, after %q", before, after)
				}
			})
		}
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := CreateRoster(t.Context(), path, bootstrapAdmin, "operator", logger); err == nil {
		t.Fatal("CreateRoster() error = nil, want a stat failure")
	} else if errors.Is(err, ErrRosterExists) {
		t.Errorf("CreateRoster() error = %v, want it to not claim the file exists", err)
	}
}

// TestAddMember_ByteIdentityOnFailure proves that a write refused for
// any reason leaves the file byte-identical: the validation failures
// caught before the lock, the collision re-check caught with the lock
// held, and the admin grant refusal are each a different code path, so
// each gets its own row. Named mutation 1 (write before validating)
// fails every row here.
func TestAddMember_ByteIdentityOnFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	const seedJSON = `{"schema_version":1,"members":[{"address":"admin@example.com","display_name":"Admin","username":"admin","admin":true},{"address":"seed@example.com","display_name":"Seed","username":"seed"}]}`

	tests := []struct {
		name      string
		candidate Member
		actor     string
		channel   identity.Channel
		wantErr   error
	}{
		{
			name:      "an address that does not parse",
			candidate: Member{Address: "not-an-address", DisplayName: "A", Username: "a"},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errInvalidMemberAddress,
		},
		{
			name:      "a duplicate address",
			candidate: Member{Address: "seed@example.com", DisplayName: "A", Username: "not-seed"},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errDuplicateAddress,
		},
		{
			name:      "a duplicate username",
			candidate: Member{Address: "new@example.com", DisplayName: "A", Username: "seed"},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errDuplicateUsername,
		},
		{
			name:      "the admin flag on the cookie channel",
			candidate: Member{Address: "new@example.com", DisplayName: "A", Username: "new", Admin: true},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errAdminGrantForbidden,
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

			err = roster.AddMember(t.Context(), tt.candidate, tt.actor, tt.channel, "203.0.113.1", logger)
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
	roster, path := newAdminRoster(t)

	candidate := Member{Address: "Alice@Example.com", DisplayName: "Alice", Username: "alice"}
	if err := roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	got, err := roster.Lookup(t.Context(), "alice@example.com")
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
// secret, or cookie value can ride along. A cookie add keeps the
// caller's client address. A command-line add or create names an
// operating-system user as its actor and reads client_ip "local"
// whatever the caller passed, so the record keeps one shape. A cookie
// actor is logged in its normalized form, the form every other record
// names that member by (security review n4). Named
// mutation 5 (drop the channel from the audit record) fails the
// "channel" key here.
func TestAddMember_AuditRecord(t *testing.T) {
	alice := Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
	tests := []struct {
		name  string
		write func(t *testing.T, logger *slog.Logger) error
		want  map[string]any
	}{
		{
			name: "a cookie add",
			write: func(t *testing.T, logger *slog.Logger) error {
				roster, _ := newAdminRoster(t)
				return roster.AddMember(t.Context(), alice, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger)
			},
			want: map[string]any{
				"actor":     "admin@example.com",
				"channel":   "cookie",
				"address":   "alice@example.com",
				"admin":     false,
				"client_ip": "198.51.100.7",
			},
		},
		{
			name: "a cookie add whose actor differs from its record in case",
			write: func(t *testing.T, logger *slog.Logger) error {
				roster, _ := newAdminRoster(t)
				return roster.AddMember(t.Context(), alice, "Admin@Example.COM", identity.ChannelCookie, "198.51.100.7", logger)
			},
			want: map[string]any{
				"actor":     "admin@example.com",
				"channel":   "cookie",
				"address":   "alice@example.com",
				"admin":     false,
				"client_ip": "198.51.100.7",
			},
		},
		{
			name: "a command-line add",
			write: func(t *testing.T, logger *slog.Logger) error {
				roster, _ := newAdminRoster(t)
				return roster.AddMember(t.Context(), alice, "operator", identity.ChannelCommandLine, "198.51.100.7", logger)
			},
			want: map[string]any{
				"actor":     "operator",
				"channel":   "command-line",
				"address":   "alice@example.com",
				"admin":     false,
				"client_ip": "local",
			},
		},
		{
			name: "a command-line create",
			write: func(t *testing.T, logger *slog.Logger) error {
				path := filepath.Join(t.TempDir(), "roster.json")
				_, err := CreateRoster(t.Context(), path, bootstrapAdmin, "operator", logger)
				return err
			},
			want: map[string]any{
				"actor":     "operator",
				"channel":   "command-line",
				"address":   "admin@example.com",
				"admin":     true,
				"client_ip": "local",
			},
		},
	}

	allowed := map[string]bool{
		"time": true, "level": true, "msg": true,
		"actor": true, "channel": true, "address": true, "admin": true, "client_ip": true,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			if err := tt.write(t, slog.New(slog.NewJSONHandler(&logBuf, nil))); err != nil {
				t.Fatalf("write: %v", err)
			}

			var record map[string]any
			if err := json.Unmarshal(bytes.TrimSpace(logBuf.Bytes()), &record); err != nil {
				t.Fatalf("unmarshal audit record %q: %v", logBuf.String(), err)
			}
			for key, wantValue := range tt.want {
				gotValue, ok := record[key]
				if !ok {
					t.Errorf("audit record missing %q", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("audit record[%q] = %v, want %v", key, gotValue, wantValue)
				}
			}
			for key, value := range record {
				if !allowed[key] {
					t.Errorf("audit record carries unexpected key %q (value %v)", key, value)
				}
			}
		})
	}
}

// TestAddMember_ConcurrentDistinctMembersBothSurvive proves that two
// concurrent adds for distinct members both succeed and the file
// carries both, under -race. Named mutation 2 (take the lock after
// the re-check) does not reach this test, since these two candidates
// never collide; TestAddMember_ConcurrentUsernameCollision
// covers that.
func TestAddMember_ConcurrentDistinctMembersBothSurvive(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	roster, path := newAdminRoster(t)

	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"},
		{Address: "bob@example.com", DisplayName: "Bob", Username: "bob"},
	}
	errs := make([]error, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate Member) {
			defer wg.Done()
			errs[i] = roster.AddMember(t.Context(), candidate, "operator", identity.ChannelCommandLine, "", logger)
		}(i, candidate)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("AddMember(%d) error = %v, want nil", i, err)
		}
	}

	// The bootstrap admin is the roster's first member, so every
	// count below is the candidates plus one.
	if got := mustList(t, roster); len(got) != len(candidates)+1 {
		t.Fatalf("List() after concurrent adds = %d members, want %d", len(got), len(candidates)+1)
	}
	onDisk, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster after concurrent adds: %v", err)
	}
	if got := mustList(t, onDisk); len(got) != len(candidates)+1 {
		t.Errorf("the file on disk carries %d members, want %d", len(got), len(candidates)+1)
	}
}

// TestAddMember_ConcurrentUsernameCollision proves that two
// concurrent adds claiming the same username leave exactly
// one member, with the other refused by ErrRosterRecordInvalid. Named
// mutation 2 (take the lock after the re-check) admits both, so this
// test fails under that mutation.
func TestAddMember_ConcurrentUsernameCollision(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	roster, _ := newAdminRoster(t)

	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice", Username: "same"},
		{Address: "bob@example.com", DisplayName: "Bob", Username: "same"},
	}
	errs := make([]error, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate Member) {
			defer wg.Done()
			errs[i] = roster.AddMember(t.Context(), candidate, "operator", identity.ChannelCommandLine, "", logger)
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
	if got := mustList(t, roster); len(got) != 2 {
		t.Fatalf("List() after the collision = %d members, want 2 (the bootstrap admin and one winner)", len(got))
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
			name:    "the retired streamer_username key alone on an otherwise valid member",
			content: `{"schema_version":1,"members":[{"address":"a@example.com","display_name":"A","streamer_username":"a"}]}`,
			wantErr: ErrRosterUnparseable,
		},
		{
			name:    "the retired streamer_username key beside username",
			content: `{"schema_version":1,"members":[{"address":"a@example.com","display_name":"A","username":"a","streamer_username":"a"}]}`,
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
			if _, err := roster.Lookup(t.Context(), "alice@example.com"); err != nil {
				t.Fatalf("Lookup(alice) before the change: %v", err)
			}

			tt.change(t, path)

			for _, row := range []struct {
				address string
				want    error
			}{{"alice@example.com", tt.wantAlice}, {"bob@example.com", tt.wantBob}} {
				_, err := roster.Lookup(t.Context(), row.address)
				if row.want == nil && err != nil {
					t.Errorf("Lookup(%s) error = %v, want nil", row.address, err)
				}
				if row.want != nil && !errors.Is(err, row.want) {
					t.Errorf("Lookup(%s) error = %v, want it to wrap %v", row.address, err, row.want)
				}
			}
			if tt.wantBob != nil {
				if _, err := roster.List(t.Context()); !errors.Is(err, tt.wantBob) {
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

			_, aliceErr := roster.Lookup(t.Context(), "alice@example.com")
			_, carolErr := roster.Lookup(t.Context(), "carol@example.com")
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
	const seed = `{"schema_version":1,"members":[{"address":"admin@example.com","display_name":"Admin","admin":true},{"address":"bob@example.com","display_name":"Bob"}]}`
	const bobRemoved = `{"schema_version":1,"members":[{"address":"admin@example.com","display_name":"Admin","admin":true},{"address":"carol@example.com","display_name":"Carol"}]}`

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

			if err := cli.AddMember(t.Context(), Member{Address: "carol@example.com", DisplayName: "Carol", Username: "carol"}, "operator", identity.ChannelCommandLine, "", logger); err != nil {
				t.Fatalf("cli AddMember: %v", err)
			}
			writeInPlace(t, path, bobRemoved)

			if lookupFirst {
				if _, err := server.Lookup(t.Context(), "bob@example.com"); !errors.Is(err, ErrMemberNotFound) {
					t.Errorf("server Lookup(bob) after the hand removal error = %v, want ErrMemberNotFound", err)
				}
			}
			if err := server.AddMember(t.Context(), Member{Address: "dave@example.com", DisplayName: "Dave", Username: "dave"}, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
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
			want := []string{"admin@example.com", "carol@example.com", "dave@example.com"}
			if !slices.Equal(got, want) {
				t.Errorf("final file members = %v, want %v", got, want)
			}
			if _, err := server.Lookup(t.Context(), "bob@example.com"); !errors.Is(err, ErrMemberNotFound) {
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
	path := writeRosterFile(t, `{"schema_version":1,"members":[{"address":"admin@example.com","display_name":"Admin","admin":true}]}`)

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
			candidate := Member{Address: fmt.Sprintf("writer%d@example.com", i), DisplayName: "Writer", Username: fmt.Sprintf("writer%d", i)}
			errs[i] = r.AddMember(t.Context(), candidate, "operator", identity.ChannelCommandLine, "", logger)
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
	if got := mustList(t, onDisk); len(got) != writers+1 {
		t.Errorf("the file carries %d members, want %d (the seeded admin plus every writer): a concurrent write was lost", len(got), writers+1)
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
				first := Member{Address: fmt.Sprintf("admin%d@example.com", i), DisplayName: "Admin", Username: fmt.Sprintf("admin%d", i), Admin: true}
				rosters[i], errs[i] = CreateRoster(t.Context(), path, first, "operator", slog.New(slog.NewTextHandler(io.Discard, nil)))
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
		writeDone <- roster.AddMember(ctx, Member{Address: "bob@example.com", DisplayName: "Bob", Username: "bob"}, "operator", identity.ChannelCommandLine, "", logger)
	}()
	select {
	case <-ctx.waiting:
	case err := <-writeDone:
		t.Fatalf("AddMember() returned %v without waiting on the held directory lock", err)
	}

	if _, err := roster.Lookup(t.Context(), "alice@example.com"); err != nil {
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

// TestRosterWrites_AdminInvariantAndGrant proves the rules every roster
// write shares: the store never writes a roster naming no admin, only
// the command line may grant the admin flag, and a cookie actor must
// hold the flag in the file as re-read under the lock. The bootstrap
// create runs every check an add does, from an absent file. Each
// refused row asserts the file's bytes, or its absence, are unchanged.
//
// The two race rows open the Roster while the admin they depend on
// still exists, then edit the file by hand before the write, so only a
// check that re-reads the disk copy explains the refusal. In the
// demoted-actor row the hand edit promotes another member, so the
// roster still names an admin and the actor check alone refuses it.
func TestRosterWrites_AdminInvariantAndGrant(t *testing.T) {
	const (
		oneAdmin = `{"schema_version":1,"members":[` +
			`{"address":"admin@example.com","display_name":"Admin","username":"admin","admin":true},` +
			`{"address":"member@example.com","display_name":"Member","username":"member"}]}`
		noAdmin = `{"schema_version":1,"members":[` +
			`{"address":"member@example.com","display_name":"Member","username":"member"}]}`
		noMembers    = `{"schema_version":1,"members":[]}`
		actorDemoted = `{"schema_version":1,"members":[` +
			`{"address":"admin@example.com","display_name":"Admin","username":"admin"},` +
			`{"address":"member@example.com","display_name":"Member","username":"member","admin":true}]}`
	)
	newAdmin := Member{Address: "new@example.com", DisplayName: "New", Username: "new", Admin: true}
	newMember := Member{Address: "new@example.com", DisplayName: "New", Username: "new"}

	tests := []struct {
		name      string
		create    bool   // CreateRoster rather than AddMember
		disk      string // the file before the write; empty means absent
		handEdit  string // when set, replaces the file after the Roster opens
		candidate Member
		channel   identity.Channel
		actor     string
		wantErr   error // nil means the write lands
	}{
		{
			name: "create: an admin from the command line", create: true,
			candidate: newAdmin, channel: identity.ChannelCommandLine, actor: "operator",
		},
		{
			name: "create: a non-admin", create: true,
			candidate: newMember, channel: identity.ChannelCommandLine, actor: "operator",
			wantErr: errNoAdmin,
		},
		{
			name: "create: an admin with no username", create: true,
			candidate: Member{Address: "new@example.com", DisplayName: "New", Admin: true},
			channel:   identity.ChannelCommandLine, actor: "operator",
			wantErr: errMissingUsername,
		},
		{
			name: "create: an admin with username .x", create: true,
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: ".x", Admin: true},
			channel:   identity.ChannelCommandLine, actor: "operator",
			wantErr: errInvalidUsername,
		},
		{
			name: "create: an admin whose display name carries U+202E", create: true,
			candidate: Member{Address: "new@example.com", DisplayName: "Ne\u202ew", Username: "new", Admin: true},
			channel:   identity.ChannelCommandLine, actor: "operator",
			wantErr: errInvalidDisplayName,
		},
		{
			name: "create: an empty actor", create: true,
			candidate: newAdmin, channel: identity.ChannelCommandLine, actor: "",
			wantErr: errInvalidActor,
		},
		{
			name: "create: the file is present", create: true, disk: oneAdmin,
			candidate: newAdmin, channel: identity.ChannelCommandLine, actor: "operator",
			wantErr: ErrRosterExists,
		},
		{
			name: "add: a non-admin to a roster naming no admin", disk: noAdmin,
			candidate: newMember, channel: identity.ChannelCommandLine, actor: "operator",
			wantErr: errNoAdmin,
		},
		{
			name: "add: a non-admin to a roster with no members", disk: noMembers,
			candidate: newMember, channel: identity.ChannelCommandLine, actor: "operator",
			wantErr: errNoAdmin,
		},
		{
			name: "add: an admin from the command line to a roster naming no admin", disk: noAdmin,
			candidate: newAdmin, channel: identity.ChannelCommandLine, actor: "operator",
		},
		{
			name: "add: an admin over a cookie by an admin actor", disk: oneAdmin,
			candidate: newAdmin, channel: identity.ChannelCookie, actor: "admin@example.com",
			wantErr: errAdminGrantForbidden,
		},
		{
			name: "add: a non-admin over a cookie by the admin", disk: oneAdmin,
			candidate: newMember, channel: identity.ChannelCookie, actor: "admin@example.com",
		},
		{
			name: "add: a non-admin over a cookie by a non-admin member", disk: oneAdmin,
			candidate: newMember, channel: identity.ChannelCookie, actor: "member@example.com",
			wantErr: errActorNotAdmin,
		},
		{
			name: "add: the actor is demoted by hand after the Roster opened", disk: oneAdmin, handEdit: actorDemoted,
			candidate: newMember, channel: identity.ChannelCookie, actor: "admin@example.com",
			wantErr: errActorNotAdmin,
		},
		{
			name: "add: the last admin is removed by hand after the Roster opened", disk: oneAdmin, handEdit: noAdmin,
			candidate: newMember, channel: identity.ChannelCommandLine, actor: "operator",
			wantErr: errNoAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			path := filepath.Join(t.TempDir(), "roster.json")
			if tt.disk != "" {
				if err := os.WriteFile(path, []byte(tt.disk), 0o600); err != nil {
					t.Fatalf("write fixture roster: %v", err)
				}
			}

			var err error
			if tt.create {
				_, err = CreateRoster(t.Context(), path, tt.candidate, tt.actor, logger)
			} else {
				roster, openErr := OpenRoster(path)
				if openErr != nil {
					t.Fatalf("OpenRoster: %v", openErr)
				}
				if tt.handEdit != "" {
					replaceRosterFile(t, path, tt.handEdit)
				}
				before := readOrAbsent(t, path)
				err = roster.AddMember(t.Context(), tt.candidate, tt.actor, tt.channel, "198.51.100.7", logger)
				if tt.wantErr != nil && !bytes.Equal(before, readOrAbsent(t, path)) {
					t.Errorf("a refused add changed the file: before %q, after %q", before, readOrAbsent(t, path))
				}
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("write error = %v, want it to wrap %v", err, tt.wantErr)
				}
				if logBuf.Len() != 0 {
					t.Errorf("a refused write logged an audit record: %q", logBuf.String())
				}
				switch {
				case tt.create && tt.disk == "":
					if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
						t.Errorf("a refused create left a file behind: Stat error = %v", statErr)
					}
				case tt.create:
					if got := readOrAbsent(t, path); !bytes.Equal(got, []byte(tt.disk)) {
						t.Errorf("a refused create changed the existing file: got %q, want %q", got, tt.disk)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("write error = %v, want nil", err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Stat after the write: %v", err)
			}
			if mode := info.Mode().Perm(); mode != 0o600 {
				t.Errorf("file mode after the write = %o, want %o", mode, 0o600)
			}
			onDisk, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster after the write: %v", err)
			}
			got, err := onDisk.Lookup(t.Context(), tt.candidate.Address)
			if err != nil {
				t.Fatalf("Lookup(candidate) after the write: %v", err)
			}
			if got.Admin != tt.candidate.Admin {
				t.Errorf("the written candidate's admin flag = %v, want %v", got.Admin, tt.candidate.Admin)
			}
			if tt.create && len(mustList(t, onDisk)) != 1 {
				t.Errorf("a created roster carries %d members, want 1", len(mustList(t, onDisk)))
			}
			if records := strings.Count(logBuf.String(), "\n"); records != 1 {
				t.Errorf("the write logged %d audit records, want 1: %q", records, logBuf.String())
			}
		})
	}
}

// readOrAbsent returns path's content, or nil when path does not exist.
func readOrAbsent(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// TestRoster_EffectivePrincipalUniqueness proves roster uniqueness keys
// on the principal the seam derives for each member: the
// username when set, otherwise the normalized address. A
// username that case-folds to another member's address, where that
// member has no username, gives both members one principal, and the
// streamer resolves a username by case-insensitive equality, so they
// would share one streamer account (security review m1). Both the
// loader and the write path refuse it, and neither error names the
// colliding value. A username naming another member's address when
// that member has a username of its own shares no principal, so it
// loads.
func TestRoster_EffectivePrincipalUniqueness(t *testing.T) {
	const (
		// legacy is a roster the loader admits: alice predates the
		// write path's username requirement and has none, so her
		// principal is her address.
		legacy = `{"schema_version":1,"members":[` +
			`{"address":"root@example.test","display_name":"Root","username":"root","admin":true},` +
			`{"address":"alice@example.test","display_name":"Alice"}]}`
		collidingValue = "alice@example.test"
	)

	loadTests := []struct {
		name    string
		members string
		wantErr error // nil means the roster loads
	}{
		{
			name: "a username equal to a later member's address, case-folded",
			members: `[{"address":"root@example.test","display_name":"Root","username":"Alice@Example.TEST","admin":true},` +
				`{"address":"alice@example.test","display_name":"Alice"}]`,
			wantErr: errDuplicatePrincipal,
		},
		{
			name: "a username equal to an earlier member's address",
			members: `[{"address":"alice@example.test","display_name":"Alice"},` +
				`{"address":"root@example.test","display_name":"Root","username":"alice@example.test","admin":true}]`,
			wantErr: errDuplicatePrincipal,
		},
		{
			name: "a username equal to its own address and to another member's username",
			members: `[{"address":"alice@example.test","display_name":"Alice","username":"alice@example.test"},` +
				`{"address":"root@example.test","display_name":"Root","username":"ALICE@example.test","admin":true}]`,
			wantErr: errDuplicateUsername,
		},
		{
			name:    "a username equal to its own address alone",
			members: `[{"address":"alice@example.test","display_name":"Alice","username":"Alice@Example.test","admin":true}]`,
		},
		{
			name: "a username equal to the address of a member who has a username",
			members: `[{"address":"alice@example.test","display_name":"Alice","username":"alice"},` +
				`{"address":"root@example.test","display_name":"Root","username":"alice@example.test","admin":true}]`,
		},
	}
	for _, tt := range loadTests {
		t.Run("load: "+tt.name, func(t *testing.T) {
			path := writeRosterFile(t, `{"schema_version":1,"members":`+tt.members+`}`)
			_, err := OpenRoster(path)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("OpenRoster error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, ErrRosterRecordInvalid) || !errors.Is(err, tt.wantErr) {
				t.Fatalf("OpenRoster error = %v, want it to wrap ErrRosterRecordInvalid and %v", err, tt.wantErr)
			}
			if strings.Contains(strings.ToLower(err.Error()), collidingValue) {
				t.Errorf("OpenRoster error = %q, want it to not name the colliding value", err.Error())
			}
		})
	}

	writeTests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{name: "a username equal to a member's address, case-folded", username: "Alice@Example.test", wantErr: errDuplicatePrincipal},
		{name: "a username equal to a member's address exactly", username: collidingValue, wantErr: errDuplicatePrincipal},
		{name: "a username equal to another member's username, case-folded", username: "ROOT", wantErr: errDuplicateUsername},
		{name: "a username naming no principal", username: "mallory"},
	}
	for _, tt := range writeTests {
		t.Run("write: "+tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			path := writeRosterFile(t, legacy)
			roster, err := OpenRoster(path)
			if err != nil {
				t.Fatalf("OpenRoster: %v", err)
			}
			candidate := Member{Address: "mallory@example.test", DisplayName: "Mallory", Username: tt.username}
			err = roster.AddMember(t.Context(), candidate, "operator", identity.ChannelCommandLine, "", logger)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("AddMember error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, ErrRosterRecordInvalid) || !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddMember error = %v, want it to wrap ErrRosterRecordInvalid and %v", err, tt.wantErr)
			}
			if strings.Contains(strings.ToLower(err.Error()), collidingValue) {
				t.Errorf("AddMember error = %q, want it to not name the colliding value", err.Error())
			}
			if got := readOrAbsent(t, path); string(got) != legacy {
				t.Errorf("a refused add changed the file: got %q, want %q", got, legacy)
			}
			if logBuf.Len() != 0 {
				t.Errorf("a refused add logged an audit record: %q", logBuf.String())
			}
		})
	}
}

// TestOpenRoster_RefusesFileWritableByOthers proves the loader refuses
// a roster file that a group or other user may write: the roster
// carries the admin flag and a changed file takes effect on the next
// request, so such a file would let another local account grant
// itself admin with no audit record (security review m3). A file only
// the owner writes loads whatever it grants for reading. The refusal
// names the path and never the file's content.
func TestOpenRoster_RefusesFileWritableByOthers(t *testing.T) {
	const content = `{"schema_version":1,"members":[{"address":"mode-fixture@example.test","display_name":"A","username":"a","admin":true}]}`
	tests := []struct {
		mode    os.FileMode
		wantErr bool
	}{
		{mode: 0o600},
		{mode: 0o640},
		{mode: 0o644},
		{mode: 0o400},
		{mode: 0o620, wantErr: true},
		{mode: 0o602, wantErr: true},
		{mode: 0o660, wantErr: true},
		{mode: 0o666, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%04o", tt.mode), func(t *testing.T) {
			path := writeRosterFile(t, content)
			if err := os.Chmod(path, tt.mode); err != nil {
				t.Fatalf("Chmod: %v", err)
			}
			_, err := OpenRoster(path)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("OpenRoster(mode %04o) error = %v, want nil", tt.mode, err)
				}
				return
			}
			if !errors.Is(err, ErrRosterUnreadable) || !errors.Is(err, errRosterWritableByOthers) {
				t.Fatalf("OpenRoster(mode %04o) error = %v, want it to wrap ErrRosterUnreadable and errRosterWritableByOthers", tt.mode, err)
			}
			if !strings.Contains(err.Error(), path) {
				t.Errorf("OpenRoster error = %q, want it to name the path", err.Error())
			}
			if strings.Contains(err.Error(), "mode-fixture@example.test") {
				t.Errorf("OpenRoster error = %q, want it to not carry the file's content", err.Error())
			}
		})
	}
}

// TestRosterWrite_LandedFailureStillAudits proves a write that
// reaches the rename logs its audit record even when a later step
// fails, because the member is on disk whatever the caller is told
// (security review m4). The failure after the rename, the directory
// fsync or the re-read, is injected through the writeRosterStream
// seam: the fsync row runs the real write and then reports an error,
// and the re-read row runs the real write and then widens the file's
// mode so the re-read refuses it. The error tells the operator the
// change landed, and the record keeps the success record's fields at
// warning level. A failure before the rename writes nothing and logs
// nothing.
func TestRosterWrite_LandedFailureStillAudits(t *testing.T) {
	errInjected := errors.New("injected failure")
	afterRenameFsync := func(path string, perm os.FileMode, write func(io.Writer) error) error {
		if err := atomicfile.WriteStream(path, perm, write); err != nil {
			return err
		}
		return errInjected
	}
	afterRenameReread := func(path string, perm os.FileMode, write func(io.Writer) error) error {
		if err := atomicfile.WriteStream(path, perm, write); err != nil {
			return err
		}
		return os.Chmod(path, 0o622)
	}
	beforeRename := func(string, os.FileMode, func(io.Writer) error) error {
		return errInjected
	}

	candidate := Member{Address: "landed@example.test", DisplayName: "Landed", Username: "landed"}
	tests := []struct {
		name       string
		create     bool
		stream     func(string, os.FileMode, func(io.Writer) error) error
		wantLanded bool
		wantAdmin  bool
		wantAddr   string
	}{
		{name: "add: the directory fsync fails after the rename", stream: afterRenameFsync, wantLanded: true, wantAddr: "landed@example.test"},
		{name: "add: the re-read fails after the rename", stream: afterRenameReread, wantLanded: true, wantAddr: "landed@example.test"},
		{name: "add: the write fails before the rename", stream: beforeRename},
		{name: "create: the directory fsync fails after the rename", create: true, stream: afterRenameFsync, wantLanded: true, wantAdmin: true, wantAddr: "admin@example.com"},
		{name: "create: the write fails before the rename", create: true, stream: beforeRename},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			var roster *Roster
			if tt.create {
				path = filepath.Join(t.TempDir(), "roster.json")
			} else {
				roster, path = newAdminRoster(t)
			}
			before := readOrAbsent(t, path)

			t.Cleanup(func() { writeRosterStream = atomicfile.WriteStream })
			writeRosterStream = tt.stream

			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			var err error
			if tt.create {
				_, err = CreateRoster(t.Context(), path, bootstrapAdmin, "operator", logger)
			} else {
				err = roster.AddMember(t.Context(), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger)
			}
			if err == nil {
				t.Fatal("write error = nil, want the injected failure reported")
			}

			if !tt.wantLanded {
				if errors.Is(err, errRosterWriteLanded) {
					t.Errorf("write error = %v, want it not to claim the change landed", err)
				}
				if logBuf.Len() != 0 {
					t.Errorf("a write that never landed logged a record: %q", logBuf.String())
				}
				if got := readOrAbsent(t, path); !bytes.Equal(got, before) {
					t.Errorf("a write that never landed changed the file: got %q, want %q", got, before)
				}
				return
			}

			if !errors.Is(err, errRosterWriteLanded) {
				t.Errorf("write error = %v, want it to wrap errRosterWriteLanded", err)
			}
			var record map[string]any
			if err := json.Unmarshal(bytes.TrimSpace(logBuf.Bytes()), &record); err != nil {
				t.Fatalf("unmarshal audit record %q: %v", logBuf.String(), err)
			}
			want := map[string]any{
				"level":   "WARN",
				"actor":   "admin@example.com",
				"channel": "cookie",
				"address": tt.wantAddr,
				"admin":   tt.wantAdmin,
			}
			if tt.create {
				want["actor"] = "operator"
				want["channel"] = "command-line"
			}
			for key, wantValue := range want {
				if got := record[key]; got != wantValue {
					t.Errorf("audit record[%q] = %v, want %v; record %q", key, got, wantValue, logBuf.String())
				}
			}
			if _, ok := record["client_ip"]; !ok {
				t.Errorf("audit record %q carries no client_ip", logBuf.String())
			}
		})
	}
}

// TestValidateWriteCandidate_BlankDisplayName proves the write path
// refuses a display name with nothing visible in it: the kind answer
// hands a stranger the admin's display name as the contact, and an
// empty one names nobody (security review n3). A name of spaces, or of
// the zero-width joiners a display name otherwise admits, is as blank
// as an empty one. The loader keeps admitting a blank name, so a
// roster written before this rule still loads.
func TestValidateWriteCandidate_BlankDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{name: "empty", displayName: "", wantErr: true},
		{name: "one space", displayName: " ", wantErr: true},
		{name: "several spaces", displayName: "   ", wantErr: true},
		{name: "a zero-width joiner alone", displayName: "\u200d", wantErr: true},
		{name: "spaces around a zero-width non-joiner", displayName: " \u200c ", wantErr: true},
		{name: "one letter", displayName: "A"},
		{name: "a letter inside spaces", displayName: " A "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := Member{Address: "a@example.com", DisplayName: tt.displayName, Username: "a"}
			_, err := validateWriteCandidate(candidate)
			if tt.wantErr != errors.Is(err, errBlankDisplayName) {
				t.Errorf("validateWriteCandidate(display name %q) error = %v, want errBlankDisplayName %v", tt.displayName, err, tt.wantErr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateWriteCandidate(display name %q) error = %v, want nil", tt.displayName, err)
			}
			if _, err := recordOf(candidate).validate(); err != nil {
				t.Errorf("validate(display name %q) error = %v, want nil: the loader keeps admitting it", tt.displayName, err)
			}
		})
	}
}

// usersBusy is the busy timeout every users-table test store opens with.
const usersBusy = 5 * time.Second

// newUsersStore creates a current store under t.TempDir() and returns
// its serve-mode handle, closed at cleanup.
func newUsersStore(t *testing.T) *store.Handle {
	t.Helper()
	h, err := store.OpenServe(context.Background(), t.TempDir(), store.ServeOptions{
		BusyTimeout:   usersBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// writeCtx returns a context carrying the deadline every store write
// needs.
func writeCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// userRow is a users row a test writes directly, before the code under
// test runs. No production path removes a member, so a tombstoned row
// exists only this way.
type userRow struct {
	id                             identity.UserID
	address, displayName, username string
	admin, tombstoned              bool
	createdAt                      int64
}

// insertUsers writes rows into h's users table in one transaction.
func insertUsers(t *testing.T, h *store.Handle, rows ...userRow) {
	t.Helper()
	err := h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
		for _, r := range rows {
			var tombstonedAt any
			if r.tombstoned {
				tombstonedAt = r.createdAt + 1
			}
			if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				r.id[:], r.address, r.displayName, r.username, r.admin, tombstonedAt, r.createdAt, r.createdAt); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert fixture users: %v", err)
	}
}

// storedUser is one users row as the table holds it.
type storedUser struct {
	id                    []byte
	displayName, username string
	admin                 int
	tombstoned            bool
}

// readUser returns address's users row, failing the test when there is
// none.
func readUser(t *testing.T, h *store.Handle, address string) storedUser {
	t.Helper()
	var u storedUser
	err := h.QueryRow(t.Context(), `SELECT id, display_name, username, admin, tombstoned_at IS NOT NULL FROM users WHERE address = ?`, address).
		Scan(&u.id, &u.displayName, &u.username, &u.admin, &u.tombstoned)
	if err != nil {
		t.Fatalf("read users row %s: %v", address, err)
	}
	return u
}

// dumpUsers renders every users row, so a refused write can be shown to
// have changed none.
func dumpUsers(t *testing.T, h *store.Handle) string {
	t.Helper()
	rows, err := h.Query(t.Context(), `SELECT hex(id), address, display_name, username, admin, coalesce(tombstoned_at, -1), created_at, updated_at FROM users ORDER BY address`)
	if err != nil {
		t.Fatalf("dump users: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var id, address, displayName, username string
		var admin, tombstonedAt, createdAt, updatedAt int64
		if err := rows.Scan(&id, &address, &displayName, &username, &admin, &tombstonedAt, &createdAt, &updatedAt); err != nil {
			t.Fatalf("scan users: %v", err)
		}
		fmt.Fprintf(&b, "%s %s %q %s %d %d %d %d\n", id, address, displayName, username, admin, tombstonedAt, createdAt, updatedAt)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("dump users: %v", err)
	}
	return b.String()
}

// Fixed ids for the users-table fixtures, so a test can assert an id
// byte for byte.
var (
	adminID  = identity.UserID{0xad, 0x01}
	aliceID  = identity.UserID{0xa1, 0x1c, 0xe0}
	goneID   = identity.UserID{0x60, 0x0e}
	formerID = identity.UserID{0xf0, 0x12}
	plainID  = identity.UserID{0x9a, 0x17}
)

// liveAdmin is a live admin row every cookie-channel fixture names as
// its actor.
var liveAdmin = userRow{id: adminID, address: "admin@example.com", displayName: "Admin", username: "groupadmin", admin: true, createdAt: 100}

// TestUsersRoster_Lookup proves the users-backed lookup: a live row
// resolves with its id, an absent or tombstoned row and an address that
// does not normalize are ErrMemberNotFound, and a store error is never
// mistaken for an unknown address. The tombstoned row fails the
// "tombstoned_at IS NULL" filter's removal.
func TestUsersRoster_Lookup(t *testing.T) {
	alice := userRow{id: aliceID, address: "alice@example.com", displayName: "Alice", username: "ally", admin: true, createdAt: 100}
	tests := []struct {
		name    string
		rows    []userRow
		address string
		close   bool
		want    Member
		wantErr error
	}{
		{
			name:    "a live row",
			rows:    []userRow{alice},
			address: "Alice@Example.COM",
			want:    Member{ID: aliceID, Address: "alice@example.com", DisplayName: "Alice", Username: "ally", Admin: true},
		},
		{name: "no row", rows: []userRow{alice}, address: "bob@example.com", wantErr: ErrMemberNotFound},
		{
			name:    "only a tombstoned row",
			rows:    []userRow{{id: goneID, address: "gone@example.com", displayName: "Gone", username: "gone", tombstoned: true, createdAt: 100}},
			address: "gone@example.com",
			wantErr: ErrMemberNotFound,
		},
		{name: "an address that does not normalize", rows: []userRow{alice}, address: "alice@example.com, bob@example.com", wantErr: ErrMemberNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUsersStore(t)
			insertUsers(t, h, tt.rows...)
			got, err := NewRoster(h).Lookup(t.Context(), tt.address)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Lookup(%q) error = %v, want %v", tt.address, err, tt.wantErr)
				}
				if strings.Contains(err.Error(), "example.com") {
					t.Errorf("Lookup error = %q, want it to name no address", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lookup(%q): %v", tt.address, err)
			}
			if got != tt.want {
				t.Errorf("Lookup(%q) = %+v, want %+v", tt.address, got, tt.want)
			}
		})
	}

	t.Run("the store closed", func(t *testing.T) {
		h := newUsersStore(t)
		insertUsers(t, h, alice)
		roster := NewRoster(h)
		if err := h.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
		_, err := roster.Lookup(t.Context(), "alice@example.com")
		if err == nil || errors.Is(err, ErrMemberNotFound) {
			t.Fatalf("Lookup on a closed store error = %v, want the store's error, not ErrMemberNotFound", err)
		}
	})
}

// TestUsersRoster_List proves List answers live rows only, in
// created_at order and then id order, whatever order they were written
// in.
func TestUsersRoster_List(t *testing.T) {
	t.Run("an empty table", func(t *testing.T) {
		got, err := NewRoster(newUsersStore(t)).List(t.Context())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("List = %+v, want none", got)
		}
	})

	t.Run("two live rows and one tombstoned row", func(t *testing.T) {
		h := newUsersStore(t)
		insertUsers(t, h,
			userRow{id: aliceID, address: "later@example.com", displayName: "Later", username: "later", createdAt: 300},
			userRow{id: goneID, address: "gone@example.com", displayName: "Gone", username: "gone", tombstoned: true, createdAt: 50},
			userRow{id: adminID, address: "first@example.com", displayName: "First", username: "first", admin: true, createdAt: 100},
		)
		got, err := NewRoster(h).List(t.Context())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		want := []Member{
			{ID: adminID, Address: "first@example.com", DisplayName: "First", Username: "first", Admin: true},
			{ID: aliceID, Address: "later@example.com", DisplayName: "Later", Username: "later"},
		}
		if !slices.Equal(got, want) {
			t.Errorf("List = %+v, want %+v", got, want)
		}
	})

	t.Run("one created_at orders by id", func(t *testing.T) {
		h := newUsersStore(t)
		insertUsers(t, h,
			userRow{id: identity.UserID{0x02}, address: "two@example.com", displayName: "Two", username: "two", admin: true, createdAt: 100},
			userRow{id: identity.UserID{0x01}, address: "one@example.com", displayName: "One", username: "one", createdAt: 100},
		)
		got, err := NewRoster(h).List(t.Context())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(got) != 2 || got[0].Address != "one@example.com" || got[1].Address != "two@example.com" {
			t.Errorf("List = %+v, want one@ then two@", got)
		}
	})
}

// TestUsersRoster_AddMember proves every write rule AddMember applies in
// its one transaction, one subtest per state. A refused row changes no
// users row. The reactivation rows assert the id byte for byte and
// re-add under the row's own old username, so a fresh id or a
// self-collision both fail them; the cookie-channel reactivation of a
// former admin fails a write that keeps the stored flag. The tombstoned
// username row answers errDuplicateUsername only when the collision
// read reaches tombstoned rows; otherwise the insert meets the UNIQUE
// constraint instead.
func TestUsersRoster_AddMember(t *testing.T) {
	gone := userRow{id: goneID, address: "gone@example.com", displayName: "Old Name", username: "gone", tombstoned: true, createdAt: 50}
	formerAdmin := userRow{id: formerID, address: "former@example.com", displayName: "Former", username: "former", admin: true, tombstoned: true, createdAt: 60}
	plain := userRow{id: plainID, address: "plain@example.com", displayName: "Plain", username: "plain", createdAt: 70}

	tests := []struct {
		name      string
		rows      []userRow
		candidate Member
		actor     string
		channel   identity.Channel
		wantErr   error
		wantText  string
		check     func(t *testing.T, h *store.Handle)
	}{
		{
			name:      "a new address and username",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "New@Example.com", DisplayName: "New", Username: "newbie"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			check: func(t *testing.T, h *store.Handle) {
				u := readUser(t, h, "new@example.com")
				if len(u.id) != len(identity.UserID{}) || bytes.Equal(u.id, make([]byte, len(u.id))) || bytes.Equal(u.id, adminID[:]) {
					t.Errorf("new row id = %x, want a fresh 16-byte id", u.id)
				}
				if u.username != "newbie" || u.displayName != "New" || u.admin != 0 || u.tombstoned {
					t.Errorf("new row = %+v, want username newbie, display New, not admin, live", u)
				}
			},
		},
		{
			name:      "a live row on the address",
			rows:      []userRow{liveAdmin, plain},
			candidate: Member{Address: "PLAIN@example.com", DisplayName: "Again", Username: "again"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errDuplicateAddress,
		},
		{
			name:      "a tombstoned row on the address, re-added with its own old username",
			rows:      []userRow{liveAdmin, gone},
			candidate: Member{Address: "gone@example.com", DisplayName: "New Name", Username: "gone"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			check: func(t *testing.T, h *store.Handle) {
				u := readUser(t, h, "gone@example.com")
				if !bytes.Equal(u.id, goneID[:]) {
					t.Errorf("reactivated id = %x, want the old id %x", u.id, goneID[:])
				}
				if u.tombstoned || u.displayName != "New Name" || u.username != "gone" {
					t.Errorf("reactivated row = %+v, want live, New Name, gone", u)
				}
			},
		},
		{
			name:      "a tombstoned admin row, re-added on the cookie channel",
			rows:      []userRow{liveAdmin, formerAdmin},
			candidate: Member{Address: "former@example.com", DisplayName: "Former", Username: "former"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			check: func(t *testing.T, h *store.Handle) {
				u := readUser(t, h, "former@example.com")
				if !bytes.Equal(u.id, formerID[:]) {
					t.Errorf("reactivated id = %x, want the old id %x", u.id, formerID[:])
				}
				if u.admin != 0 || u.tombstoned {
					t.Errorf("reactivated row = %+v, want admin 0 and live", u)
				}
			},
		},
		{
			name:      "another live row's username, differing only in ASCII case",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "GroupAdmin"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errDuplicateUsername,
		},
		{
			name:      "a new address asking for a tombstoned row's username, differing in ASCII case",
			rows:      []userRow{liveAdmin, gone},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "GONE"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errDuplicateUsername,
		},
		{
			name:      "no username",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New"},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errMissingUsername,
		},
		{
			name:      "admin requested on the cookie channel",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie", Admin: true},
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errAdminGrantForbidden,
		},
		{
			name:      "admin on the command-line channel",
			candidate: Member{Address: "first@example.com", DisplayName: "First", Username: "first", Admin: true},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			check: func(t *testing.T, h *store.Handle) {
				if u := readUser(t, h, "first@example.com"); u.admin != 1 || u.tombstoned {
					t.Errorf("command-line admin row = %+v, want admin 1 and live", u)
				}
			},
		},
		{
			name:      "a cookie actor with no live row",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			actor:     "stranger@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errActorNotAdmin,
		},
		{
			name:      "a cookie actor with a non-admin live row",
			rows:      []userRow{liveAdmin, plain},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			actor:     "plain@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errActorNotAdmin,
		},
		{
			name:      "a cookie actor with only a tombstoned admin row",
			rows:      []userRow{liveAdmin, formerAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			actor:     "former@example.com",
			channel:   identity.ChannelCookie,
			wantErr:   errActorNotAdmin,
		},
		{
			name:      "a command-line non-admin add into an empty table",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errNoAdmin,
			wantText:  "member add --admin",
		},
		{
			name:      "a command-line non-admin add beside only a tombstoned admin",
			rows:      []userRow{formerAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errNoAdmin,
			wantText:  "member add --admin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUsersStore(t)
			insertUsers(t, h, tt.rows...)
			before := dumpUsers(t, h)

			err := NewRoster(h).AddMember(writeCtx(t), tt.candidate, tt.actor, tt.channel, "192.0.2.1", slog.New(slog.DiscardHandler))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("AddMember error = %v, want %v", err, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantText) {
					t.Errorf("AddMember error = %q, want it to contain %q", err, tt.wantText)
				}
				if after := dumpUsers(t, h); after != before {
					t.Errorf("a refused add changed the users table:\nbefore %s\nafter  %s", before, after)
				}
				return
			}
			if err != nil {
				t.Fatalf("AddMember: %v", err)
			}
			tt.check(t, h)
		})
	}
}
