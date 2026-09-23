package auth

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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
	}{
		{
			name:    "a duplicate address differing only in case",
			members: `[{"address":"Alice@Example.com","display_name":"A"},{"address":"alice@example.com","display_name":"B"}]`,
			wantErr: true,
		},
		{
			name:    "a duplicate streamer_username",
			members: `[{"address":"a@example.com","display_name":"A","streamer_username":"same"},{"address":"b@example.com","display_name":"B","streamer_username":"same"}]`,
			wantErr: true,
		},
		{
			name:    "a control character in a display name",
			members: `[{"address":"a@example.com","display_name":"A\u0007B"}]`,
			wantErr: true,
		},
		{
			name:    "a control character in an address",
			members: `[{"address":"a\u0007b@example.com","display_name":"A"}]`,
			wantErr: true,
		},
		{
			name:    "an unparseable address",
			members: `[{"address":"not-an-address","display_name":"A"}]`,
			wantErr: true,
		},
		{
			name:    "display name exactly at the length cap",
			members: `[{"address":"a@example.com","display_name":"` + atCapDisplay + `"}]`,
			wantErr: false,
		},
		{
			name:    "display name one byte over the length cap",
			members: `[{"address":"a@example.com","display_name":"` + overCapDisplay + `"}]`,
			wantErr: true,
		},
		{
			name:    "streamer_username exactly at the length cap",
			members: `[{"address":"a@example.com","display_name":"A","streamer_username":"` + atCapUsername + `"}]`,
			wantErr: false,
		},
		{
			name:    "streamer_username one byte over the length cap",
			members: `[{"address":"a@example.com","display_name":"A","streamer_username":"` + overCapUsername + `"}]`,
			wantErr: true,
		},
		{
			name:    "a streamer_username carrying a non-ASCII-printable byte",
			members: `[{"address":"a@example.com","display_name":"A","streamer_username":"na` + "é" + `me"}]`,
			wantErr: true,
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
			if tt.wantErr && !errors.Is(err, ErrRosterRecordInvalid) {
				t.Errorf("OpenRoster() error = %v, want it to wrap ErrRosterRecordInvalid", err)
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

// TestNormalizeMemberAddress_SingleResolutionPoint proves R1's single
// resolution point by grep: normalizeMemberAddress is the only
// function in this file that calls identity.NormalizeAddress. Named
// mutation 5 (add a second address-to-identity derivation) fails this
// test.
func TestNormalizeMemberAddress_SingleResolutionPoint(t *testing.T) {
	source, err := os.ReadFile("roster.go")
	if err != nil {
		t.Fatalf("read roster.go: %v", err)
	}
	matches := regexp.MustCompile(`identity\.NormalizeAddress\(`).FindAllIndex(source, -1)
	if len(matches) != 1 {
		t.Errorf("roster.go calls identity.NormalizeAddress %d times, want exactly 1 (normalizeMemberAddress alone)", len(matches))
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

	if _, err := CreateRoster(path); err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("CreateRoster() file mode = %o, want %o", mode, 0o600)
	}

	if _, err := CreateRoster(path); !errors.Is(err, ErrRosterExists) {
		t.Fatalf("CreateRoster(existing) error = %v, want ErrRosterExists", err)
	}

	loaded, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster(after create) error = %v, want nil: the read-open must still work once the file exists", err)
	}
	if got := loaded.List(); len(got) != 0 {
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
	if len(roster.List()) == 0 {
		t.Error("config/roster.example.json parsed to zero members, want at least one")
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

// TestContainsControl is a direct table-driven test of the helper.
func TestContainsControl(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "plain text", input: "Example Name", want: false},
		{name: "bell character", input: "a\u0007b", want: true},
		{name: "DEL", input: "a\u007fb", want: true},
		{name: "newline", input: "a\nb", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsControl(tt.input); got != tt.want {
				t.Errorf("containsControl(%q) = %v, want %v", tt.input, got, tt.want)
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
	if _, err := CreateRoster(path); err == nil {
		t.Fatal("CreateRoster() error = nil, want a stat failure")
	} else if errors.Is(err, ErrRosterExists) {
		t.Errorf("CreateRoster() error = %v, want it to not claim the file exists", err)
	}
}
