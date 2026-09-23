package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

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
			name:    "a C1 control character (CSI) in a display name",
			members: `[{"address":"a@example.com","display_name":"A\u009bB"}]`,
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
		{name: "C1 CSI", input: "a\u009bb", want: true},
		{name: "non-control non-ASCII", input: "Zoë", want: false},
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
			candidate: Member{Address: "seed@example.com", DisplayName: "A"},
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

			err = roster.AddMember(tt.candidate, "tester", tt.channel, "203.0.113.1", logger)
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
	roster, err := CreateRoster(path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidate := Member{Address: "Alice@Example.com", DisplayName: "Alice", StreamerUsername: "alice"}
	if err := roster.AddMember(candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
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
	roster, err := CreateRoster(path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidate := Member{Address: "alice@example.com", DisplayName: "Alice"}
	if err := roster.AddMember(candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", logger); err != nil {
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
	roster, err := CreateRoster(path)
	if err != nil {
		t.Fatalf("CreateRoster: %v", err)
	}

	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice"},
		{Address: "bob@example.com", DisplayName: "Bob"},
	}
	errs := make([]error, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate Member) {
			defer wg.Done()
			errs[i] = roster.AddMember(candidate, "tester", identity.ChannelServiceToken, "203.0.113.1", logger)
		}(i, candidate)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("AddMember(%d) error = %v, want nil", i, err)
		}
	}

	if got := roster.List(); len(got) != len(candidates) {
		t.Fatalf("List() after concurrent adds = %d members, want %d", len(got), len(candidates))
	}
	onDisk, err := OpenRoster(path)
	if err != nil {
		t.Fatalf("OpenRoster after concurrent adds: %v", err)
	}
	if got := onDisk.List(); len(got) != len(candidates) {
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
	roster, err := CreateRoster(path)
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
			errs[i] = roster.AddMember(candidate, "tester", identity.ChannelServiceToken, "203.0.113.1", logger)
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
	if got := roster.List(); len(got) != 1 {
		t.Fatalf("List() after the collision = %d members, want 1", len(got))
	}
}
