package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
)

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

// TestValidateWriteCandidate_UsernameShape proves the write
// path's narrower shape check: an alphanumeric first character, then
// any run of alphanumerics, '.', '_', '@', '+', or '-', within the
// length bound.
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
		{name: "an emoji with its presentation selector (admitted)", input: "Alice \u2764\ufe0f", want: false},
		{name: "a Mongolian free variation selector (admitted)", input: "a\u180bb", want: false},
		{name: "Hangul filler U+3164 (Lo, derived set)", input: "a\u3164b", want: true},
		{name: "halfwidth Hangul filler U+FFA0 (Lo, derived set)", input: "a\uffa0b", want: true},
		{name: "combining grapheme joiner U+034F (Mn, derived set)", input: "a\u034fb", want: true},
		{name: "braille pattern blank U+2800", input: "a\u2800b", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayNameCarriesRefusedRune(tt.input); got != tt.want {
				t.Errorf("displayNameCarriesRefusedRune(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// defaultIgnorableRoster is the same ten code points
// internal/identity's TestNormalizeAddress_DefaultIgnorableRefused
// pins, checked here against AddMember on both channels. The three
// preExisting rows were already refused, before the derived-set refusal
// existed, by AddMember's reliance on identity.NormalizeAddress, which
// already refused the whole Cf category; they are labelled regressions
// rather than proofs of the derived set's own reach.
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
// r between two letters, the same shape TestValidateWriteCandidate_Fields
// uses for its own invisible-rune rows.
func defaultIgnorableAddress(r rune) string {
	return "\"a" + string(r) + "b\"@example.com"
}

// newRosterOver returns a roster over a fresh store holding rows. A row
// with no id, username, or creation stamp gets one, so a fixture names
// only what the test reads.
func newRosterOver(t *testing.T, rows ...userRow) (*Roster, *store.Handle) {
	t.Helper()
	h := newUsersStore(t)
	for i := range rows {
		if rows[i].id == (identity.UserID{}) {
			rows[i].id = identity.UserID{0xf1, byte(i + 1)}
		}
		if rows[i].username == "" {
			rows[i].username, _, _ = strings.Cut(rows[i].address, "@")
		}
		if rows[i].createdAt == 0 {
			rows[i].createdAt = int64(100 + i)
		}
	}
	insertUsers(t, h, rows...)
	return NewRoster(h), h
}

// TestRoster_LookupErrorNamesNoAddress proves a Lookup refusal for an
// input net/mail cannot parse carries neither address in it: net/mail
// quotes the unparsed remainder of a two-address list, so wrapping its
// error would echo the second address back to the caller.
func TestRoster_LookupErrorNamesNoAddress(t *testing.T) {
	const first = "lookup-first@example.com"
	const second = "lookup-second@example.com"
	roster, _ := newRosterOver(t)

	_, err := roster.Lookup(t.Context(), first+", "+second)
	if !errors.Is(err, errMemberNotFound) {
		t.Fatalf("Lookup() error = %v, want it to wrap errMemberNotFound", err)
	}
	for _, addr := range []string{first, second} {
		if strings.Contains(err.Error(), addr) {
			t.Errorf("Lookup() error = %q, want it to not contain %q", err.Error(), addr)
		}
	}
}

// TestValidateWriteCandidate_Fields proves the write path's field list,
// each row with its own written-out expectation.
func TestValidateWriteCandidate_Fields(t *testing.T) {
	// atCapDisplay and overCapDisplay pad with "é" (a 2-byte, 1-rune
	// UTF-8 character) so a byte-length cap and a rune-count cap
	// disagree: 100 of them lands at 200 bytes (maxDisplayNameLength)
	// but 100 runes, and one more crosses the byte cap while staying at
	// 101 runes, far under a rune-based cap of 200.
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
		name        string
		address     string
		displayName string
		username    string
		wantErr     error // nil means the candidate validates
	}{
		{name: "a control character in a display name", address: "a@example.com", displayName: "A\u0007B", username: "a", wantErr: errInvalidDisplayName},
		{name: "a C1 control character (CSI) in a display name", address: "a@example.com", displayName: "A\u009bB", username: "a", wantErr: errInvalidDisplayName},
		{name: "a no-break space in a display name", address: "a@example.com", displayName: "A B", username: "a", wantErr: errInvalidDisplayName},
		{name: "a right-to-left override in a display name", address: "a@example.com", displayName: "A\u202eB", username: "a", wantErr: errInvalidDisplayName},
		{name: "a zero-width space in a display name", address: "a@example.com", displayName: "A\u200bB", username: "a", wantErr: errInvalidDisplayName},
		{name: "a zero-width non-joiner in a display name", address: "a@example.com", displayName: "A\u200cB", username: "a"},
		{name: "a zero-width joiner in a display name", address: "a@example.com", displayName: "A\u200dB", username: "a"},
		{name: "a control character in an address", address: "a\u0007b@example.com", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "an unparseable address", address: "not-an-address", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "a NEL control in a quoted address local part", address: "\"a\u0085b\"@example.com", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "a C1 CSI control in a quoted address local part", address: "\"a\u009bb\"@example.com", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "a right-to-left override in a quoted address local part", address: "\"a\u202eb\"@example.com", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "a no-break space in a quoted address local part", address: "\"a b\"@example.com", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "a zero-width space in a quoted address local part", address: "\"a\u200bb\"@example.com", displayName: "A", username: "a", wantErr: errInvalidMemberAddress},
		{name: "display name exactly at the length cap", address: "a@example.com", displayName: atCapDisplay, username: "a"},
		{name: "display name one byte over the length cap", address: "a@example.com", displayName: overCapDisplay, username: "a", wantErr: errInvalidDisplayName},
		{name: "username exactly at the length cap", address: "a@example.com", displayName: "A", username: atCapUsername},
		{name: "username one byte over the length cap", address: "a@example.com", displayName: "A", username: overCapUsername, wantErr: errInvalidUsername},
		{name: "a username carrying a non-ASCII byte", address: "a@example.com", displayName: "A", username: "naéme", wantErr: errInvalidUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateWriteCandidate(Member{Address: tt.address, DisplayName: tt.displayName, Username: tt.username})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("validateWriteCandidate() error = %v, want it to wrap %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateWriteCandidate() error = %v, want nil", err)
			}
			if got.Address != tt.address || got.DisplayName != tt.displayName || got.Username != tt.username {
				t.Errorf("validateWriteCandidate() = %+v, want the candidate's own fields", got)
			}
		})
	}
}

// TestValidateWriteCandidate_UsernameRequired proves the write path
// refuses a candidate with no username.
func TestValidateWriteCandidate_UsernameRequired(t *testing.T) {
	candidate := Member{Address: "a@example.com", DisplayName: "A"}
	if _, err := validateWriteCandidate(candidate); !errors.Is(err, errMissingUsername) {
		t.Errorf("validateWriteCandidate(no username) error = %v, want it to wrap errMissingUsername", err)
	}
}

// TestAddMember_ActorAndChannelRefusals proves AddMember refuses
// before any write when the channel is not one Channel.Valid accepts,
// or when the actor does not meet its channel's own rule: a cookie
// actor must normalize as a login address, and a command-line actor
// must be non-empty and carry no rune an address refuses. Fixture
// state: a table holding a live admin, so the refusal is the actor's and
// not the missing admin's.
func TestAddMember_ActorAndChannelRefusals(t *testing.T) {
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
		{name: "a command-line actor carrying Hangul filler U+3164", actor: "oper\u3164ator", channel: identity.ChannelCommandLine},
		{name: "a command-line actor carrying braille pattern blank U+2800", actor: "operator\u2800", channel: identity.ChannelCommandLine},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roster, h := newRosterOver(t, liveAdmin)
			before := dumpUsers(t, h)

			err := roster.AddMember(writeCtx(t), candidate, tt.actor, tt.channel, "198.51.100.7", slog.New(slog.DiscardHandler))
			if !errors.Is(err, errInvalidActor) {
				t.Errorf("AddMember() error = %v, want it to wrap errInvalidActor", err)
			}
			if !errors.Is(err, errRosterRecordInvalid) {
				t.Errorf("AddMember() error = %v, want it to also wrap errRosterRecordInvalid", err)
			}
			if after := dumpUsers(t, h); after != before {
				t.Errorf("AddMember() on a refused actor or channel changed the users table:\nbefore %s\nafter  %s", before, after)
			}
		})
	}
}

// TestAddMember_ZeroWidthAddressVariantRefused proves that adding an
// address carrying a zero-width character is refused outright, so it
// can never land beside an existing member as a visually
// indistinguishable second entry. Fixture state: a table already
// holding the plain address.
func TestAddMember_ZeroWidthAddressVariantRefused(t *testing.T) {
	roster, h := newRosterOver(t, liveAdmin, userRow{address: "alice@example.com", displayName: "Alice", username: "alice"})
	before := dumpUsers(t, h)

	candidate := Member{Address: "\"a\u200dlice\"@example.com", DisplayName: "Alice", Username: "alice-2"}
	err := roster.AddMember(writeCtx(t), candidate, "admin@example.com", identity.ChannelCookie, "198.51.100.7", slog.New(slog.DiscardHandler))
	if !errors.Is(err, errRosterRecordInvalid) {
		t.Errorf("AddMember(zero-width address variant) error = %v, want it to wrap errRosterRecordInvalid", err)
	}
	if !errors.Is(err, errInvalidMemberAddress) {
		t.Errorf("AddMember(zero-width address variant) error = %v, want it to wrap errInvalidMemberAddress", err)
	}
	if after := dumpUsers(t, h); after != before {
		t.Errorf("AddMember(zero-width address variant) changed the users table:\nbefore %s\nafter  %s", before, after)
	}
}

// TestAddMember_DefaultIgnorableAddressRefused proves the derived-set
// refusal at AddMember, on both channels: a new member's address
// carrying one of the ten code points is refused and changes no row.
// Fixture state: the same address with no code point is accepted by
// AddMember on both channels, proved once each, so the refusal below is
// the code point's.
func TestAddMember_DefaultIgnorableAddressRefused(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	for _, tt := range []struct {
		channel  identity.Channel
		actor    string
		address  string
		username string
	}{
		{identity.ChannelCookie, "admin@example.com", "ab@example.com", "a"},
		{identity.ChannelCommandLine, "operator", "cd@example.com", "c"},
	} {
		roster, _ := newRosterOver(t, liveAdmin)
		plain := Member{Address: tt.address, DisplayName: "A", Username: tt.username}
		if err := roster.AddMember(writeCtx(t), plain, tt.actor, tt.channel, "198.51.100.7", logger); err != nil {
			t.Fatalf("AddMember(plain fixture, %s) = %v, want nil", tt.channel, err)
		}
	}

	for _, tt := range defaultIgnorableRoster {
		for _, channel := range []identity.Channel{identity.ChannelCookie, identity.ChannelCommandLine} {
			t.Run(tt.name+"/"+channel.String(), func(t *testing.T) {
				roster, h := newRosterOver(t, liveAdmin)
				before := dumpUsers(t, h)

				actor := "admin@example.com"
				if channel == identity.ChannelCommandLine {
					actor = "operator"
				}
				candidate := Member{Address: defaultIgnorableAddress(tt.r), DisplayName: "A", Username: "a"}
				err := roster.AddMember(writeCtx(t), candidate, actor, channel, "198.51.100.7", logger)
				if err == nil {
					t.Fatalf("AddMember() error = nil, want a refusal")
				}
				if !errors.Is(err, errRosterRecordInvalid) {
					t.Errorf("AddMember() error = %v, want it to wrap errRosterRecordInvalid", err)
				}
				if !errors.Is(err, errInvalidMemberAddress) {
					t.Errorf("AddMember() error = %v, want it to wrap errInvalidMemberAddress", err)
				}
				if after := dumpUsers(t, h); after != before {
					t.Errorf("AddMember() on a refused address changed the users table:\nbefore %s\nafter  %s", before, after)
				}
			})
		}
	}
}

// TestAddMember_AuditRecord proves the audit record AddMember writes on
// each channel: exactly the five fields the record names beside the
// logger's own, the cookie actor normalized, and the command-line
// client address read as "local" whatever the caller passed.
func TestAddMember_AuditRecord(t *testing.T) {
	alice := Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
	tests := []struct {
		name      string
		candidate Member
		actor     string
		channel   identity.Channel
		want      map[string]any
	}{
		{
			name:      "a cookie add",
			candidate: alice,
			actor:     "admin@example.com",
			channel:   identity.ChannelCookie,
			want: map[string]any{
				"actor": "admin@example.com", "channel": "cookie", "address": "alice@example.com",
				"admin": false, "client_ip": "198.51.100.7",
			},
		},
		{
			name:      "a cookie add whose actor differs from its record in case",
			candidate: alice,
			actor:     "Admin@Example.COM",
			channel:   identity.ChannelCookie,
			want: map[string]any{
				"actor": "admin@example.com", "channel": "cookie", "address": "alice@example.com",
				"admin": false, "client_ip": "198.51.100.7",
			},
		},
		{
			name:      "a command-line add",
			candidate: alice,
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			want: map[string]any{
				"actor": "operator", "channel": "command-line", "address": "alice@example.com",
				"admin": false, "client_ip": "local",
			},
		},
		{
			name:      "a command-line admin add",
			candidate: Member{Address: "boss@example.com", DisplayName: "Boss", Username: "boss", Admin: true},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			want: map[string]any{
				"actor": "operator", "channel": "command-line", "address": "boss@example.com",
				"admin": true, "client_ip": "local",
			},
		},
	}
	allowed := map[string]bool{
		"time": true, "level": true, "msg": true,
		"actor": true, "channel": true, "address": true, "admin": true, "client_ip": true,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roster, _ := newRosterOver(t, liveAdmin)
			var logBuf bytes.Buffer
			err := roster.AddMember(writeCtx(t), tt.candidate, tt.actor, tt.channel, "198.51.100.7", slog.New(slog.NewJSONHandler(&logBuf, nil)))
			if err != nil {
				t.Fatalf("AddMember: %v", err)
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

// addConcurrently runs one AddMember per candidate on the command line
// at once and returns each call's error, in candidate order.
func addConcurrently(t *testing.T, roster *Roster, candidates []Member) []error {
	t.Helper()
	errs := make([]error, len(candidates))
	var wg sync.WaitGroup
	for i, candidate := range candidates {
		wg.Add(1)
		go func(i int, candidate Member) {
			defer wg.Done()
			errs[i] = roster.AddMember(writeCtx(t), candidate, "operator", identity.ChannelCommandLine, "", slog.New(slog.DiscardHandler))
		}(i, candidate)
	}
	wg.Wait()
	return errs
}

// TestAddMember_ConcurrentDistinctMembersBothSurvive proves that two
// concurrent adds for distinct members both succeed and the table
// carries both, under -race.
func TestAddMember_ConcurrentDistinctMembersBothSurvive(t *testing.T) {
	roster, _ := newRosterOver(t, liveAdmin)
	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"},
		{Address: "bob@example.com", DisplayName: "Bob", Username: "bob"},
	}
	for i, err := range addConcurrently(t, roster, candidates) {
		if err != nil {
			t.Errorf("AddMember(%d) error = %v, want nil", i, err)
		}
	}

	// The live admin is the table's first row, so the count is the
	// candidates plus one.
	got, err := roster.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != len(candidates)+1 {
		t.Errorf("List() after concurrent adds = %d members, want %d", len(got), len(candidates)+1)
	}
}

// TestAddMember_ConcurrentUsernameCollision proves that two concurrent
// adds claiming the same username leave exactly one member, with the
// other refused by errDuplicateUsername. The collision read runs
// inside the write transaction, so a read taken outside it admits both
// and fails this test.
func TestAddMember_ConcurrentUsernameCollision(t *testing.T) {
	roster, _ := newRosterOver(t, liveAdmin)
	candidates := []Member{
		{Address: "alice@example.com", DisplayName: "Alice", Username: "same"},
		{Address: "bob@example.com", DisplayName: "Bob", Username: "same"},
	}
	successes, refusals := 0, 0
	for _, err := range addConcurrently(t, roster, candidates) {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, errDuplicateUsername):
			refusals++
		default:
			t.Errorf("AddMember() error = %v, want nil or errDuplicateUsername", err)
		}
	}
	if successes != 1 || refusals != 1 {
		t.Fatalf("got %d successes and %d refusals, want exactly one of each", successes, refusals)
	}
	got, err := roster.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() after the collision = %d members, want 2 (the live admin and one winner)", len(got))
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
		{name: "a variation selector alone", displayName: "\ufe0f", wantErr: true},
		{name: "an emoji with its presentation selector", displayName: "\u2764\ufe0f"},
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

// insertShortIDUser writes a live users row whose id is three bytes,
// which the schema's BLOB column accepts and a UserID cannot hold.
func insertShortIDUser(t *testing.T, h *store.Handle, address string) {
	t.Helper()
	err := h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, created_at, updated_at)
			VALUES (?, ?, 'Short', 'short', 0, 100, 100)`, []byte{0x01, 0x02, 0x03}, address)
		return err
	})
	if err != nil {
		t.Fatalf("insert short-id user: %v", err)
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
// does not normalize are errMemberNotFound, and a store error is never
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
		{name: "no row", rows: []userRow{alice}, address: "bob@example.com", wantErr: errMemberNotFound},
		{
			name:    "only a tombstoned row",
			rows:    []userRow{{id: goneID, address: "gone@example.com", displayName: "Gone", username: "gone", tombstoned: true, createdAt: 100}},
			address: "gone@example.com",
			wantErr: errMemberNotFound,
		},
		{name: "an address that does not normalize", rows: []userRow{alice}, address: "alice@example.com, bob@example.com", wantErr: errMemberNotFound},
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
		if err == nil || errors.Is(err, errMemberNotFound) {
			t.Fatalf("Lookup on a closed store error = %v, want the store's error, not errMemberNotFound", err)
		}
	})

	t.Run("a stored id that is not 16 bytes", func(t *testing.T) {
		h := newUsersStore(t)
		insertShortIDUser(t, h, "short@example.com")
		got, err := NewRoster(h).Lookup(t.Context(), "short@example.com")
		if err == nil || errors.Is(err, errMemberNotFound) {
			t.Fatalf("Lookup of a short-id row = %+v, %v; want an error that is not errMemberNotFound", got, err)
		}
		if !strings.Contains(err.Error(), "3 bytes") {
			t.Errorf("Lookup error = %q, want it to name the 3-byte id", err)
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

	// A store error must surface, never read as an empty roster: the
	// admin contact and the roster API build on List.
	t.Run("the store closed", func(t *testing.T) {
		h := newUsersStore(t)
		insertUsers(t, h, liveAdmin)
		roster := NewRoster(h)
		if err := h.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
		got, err := roster.List(t.Context())
		if err == nil || got != nil {
			t.Fatalf("List on a closed store = %+v, %v; want nil and the store's error", got, err)
		}
	})

	t.Run("a stored id that is not 16 bytes", func(t *testing.T) {
		h := newUsersStore(t)
		insertUsers(t, h, liveAdmin)
		insertShortIDUser(t, h, "short@example.com")
		got, err := NewRoster(h).List(t.Context())
		if err == nil || got != nil {
			t.Fatalf("List with a short-id row = %+v, %v; want nil and an error", got, err)
		}
		if !strings.Contains(err.Error(), "3 bytes") {
			t.Errorf("List error = %q, want it to name the 3-byte id", err)
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
			name:      "an admin with no username",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Admin: true},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errMissingUsername,
		},
		{
			name:      "an admin with username .x",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: ".x", Admin: true},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errInvalidUsername,
		},
		{
			name:      "an admin whose display name carries U+202E",
			candidate: Member{Address: "new@example.com", DisplayName: "Ne\u202ew", Username: "new", Admin: true},
			actor:     "operator",
			channel:   identity.ChannelCommandLine,
			wantErr:   errInvalidDisplayName,
		},
		{
			name:      "an empty command-line actor",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "new", Admin: true},
			actor:     "",
			channel:   identity.ChannelCommandLine,
			wantErr:   errInvalidActor,
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

			var logBuf bytes.Buffer
			err := NewRoster(h).AddMember(writeCtx(t), tt.candidate, tt.actor, tt.channel, "192.0.2.1", slog.New(slog.NewJSONHandler(&logBuf, nil)))
			if tt.wantErr != nil {
				if logBuf.Len() != 0 {
					t.Errorf("a refused add logged an audit record: %q", logBuf.String())
				}
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
			if records := strings.Count(logBuf.String(), "\n"); records != 1 {
				t.Errorf("the add logged %d audit records, want 1: %q", records, logBuf.String())
			}
			tt.check(t, h)
		})
	}
}
