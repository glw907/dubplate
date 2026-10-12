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
// function in this file that resolves a roster member's address. A
// write's actor is named by id, so no other function has an address to
// resolve. Named mutation 5 (add a second member-address derivation
// outside normalizeMemberAddress) fails this test.
func TestNormalizeMemberAddress_SingleResolutionPoint(t *testing.T) {
	source, err := os.ReadFile("roster.go")
	if err != nil {
		t.Fatalf("read roster.go: %v", err)
	}
	calls := regexp.MustCompile(`identity\.NormalizeAddress\(`)
	body := regexp.MustCompile(`(?s)func normalizeMemberAddress\(.*?\n}`).Find(source)
	if body == nil {
		t.Fatal("roster.go has no func normalizeMemberAddress")
	}
	if n := len(calls.FindAllIndex(body, -1)); n != 1 {
		t.Errorf("normalizeMemberAddress calls identity.NormalizeAddress %d times, want exactly 1", n)
	}
	if total := len(calls.FindAllIndex(source, -1)); total != 1 {
		t.Errorf("roster.go calls identity.NormalizeAddress %d times, want exactly 1 (normalizeMemberAddress)", total)
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
				if !errors.Is(err, ErrInvalidUsername) {
					t.Errorf("validateWriteCandidate(%q) error = %v, want it to wrap ErrInvalidUsername", tt.want, err)
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
	return NewRoster(h, newEpochStore(t, t.TempDir()), fixedNow(rosterNow)), h
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
	if !errors.Is(err, ErrMemberNotFound) {
		t.Fatalf("Lookup() error = %v, want it to wrap ErrMemberNotFound", err)
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
		{name: "a control character in a display name", address: "a@example.com", displayName: "A\u0007B", username: "a", wantErr: ErrInvalidDisplayName},
		{name: "a C1 control character (CSI) in a display name", address: "a@example.com", displayName: "A\u009bB", username: "a", wantErr: ErrInvalidDisplayName},
		{name: "a no-break space in a display name", address: "a@example.com", displayName: "A B", username: "a", wantErr: ErrInvalidDisplayName},
		{name: "a right-to-left override in a display name", address: "a@example.com", displayName: "A\u202eB", username: "a", wantErr: ErrInvalidDisplayName},
		{name: "a zero-width space in a display name", address: "a@example.com", displayName: "A\u200bB", username: "a", wantErr: ErrInvalidDisplayName},
		{name: "a zero-width non-joiner in a display name", address: "a@example.com", displayName: "A\u200cB", username: "a"},
		{name: "a zero-width joiner in a display name", address: "a@example.com", displayName: "A\u200dB", username: "a"},
		{name: "a control character in an address", address: "a\u0007b@example.com", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "an unparseable address", address: "not-an-address", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "a NEL control in a quoted address local part", address: "\"a\u0085b\"@example.com", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "a C1 CSI control in a quoted address local part", address: "\"a\u009bb\"@example.com", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "a right-to-left override in a quoted address local part", address: "\"a\u202eb\"@example.com", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "a no-break space in a quoted address local part", address: "\"a b\"@example.com", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "a zero-width space in a quoted address local part", address: "\"a\u200bb\"@example.com", displayName: "A", username: "a", wantErr: ErrInvalidMemberAddress},
		{name: "display name exactly at the length cap", address: "a@example.com", displayName: atCapDisplay, username: "a"},
		{name: "display name one byte over the length cap", address: "a@example.com", displayName: overCapDisplay, username: "a", wantErr: ErrInvalidDisplayName},
		{name: "username exactly at the length cap", address: "a@example.com", displayName: "A", username: atCapUsername},
		{name: "username one byte over the length cap", address: "a@example.com", displayName: "A", username: overCapUsername, wantErr: ErrInvalidUsername},
		{name: "a username carrying a non-ASCII byte", address: "a@example.com", displayName: "A", username: "naéme", wantErr: ErrInvalidUsername},
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
	if _, err := validateWriteCandidate(candidate); !errors.Is(err, ErrMissingUsername) {
		t.Errorf("validateWriteCandidate(no username) error = %v, want it to wrap ErrMissingUsername", err)
	}
}

// TestAddMember_ActorAndChannelRefusals proves AddMember refuses
// before any write when the actor names no channel, or does not meet
// its channel's own rule: a member actor must carry a non-zero id, and
// a command-line actor must be non-empty and carry no rune an address
// refuses. Fixture
// state: a table holding a live admin, so the refusal is the actor's and
// not the missing admin's.
func TestAddMember_ActorAndChannelRefusals(t *testing.T) {
	candidate := Member{Address: "new@example.com", DisplayName: "New", Username: "new"}
	tests := []struct {
		name string
		by   Actor
	}{
		{name: "the zero actor", by: Actor{}},
		{name: "a member actor with the zero id", by: MemberActor(identity.UserID{}, "198.51.100.7")},
		{name: "an empty actor on the command line", by: OperatorActor("")},
		{name: "a command-line actor carrying a space", by: OperatorActor("some operator")},
		{name: "a command-line actor carrying U+202E", by: OperatorActor("oper\u202eator")},
		{name: "a command-line actor carrying a newline", by: OperatorActor("operator\nforged")},
		{name: "a command-line actor carrying Hangul filler U+3164", by: OperatorActor("oper\u3164ator")},
		{name: "a command-line actor carrying braille pattern blank U+2800", by: OperatorActor("operator\u2800")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roster, h := newRosterOver(t, liveAdmin)
			before := dumpUsers(t, h)

			err := roster.AddMember(writeCtx(t), candidate, tt.by, slog.New(slog.DiscardHandler))
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
	err := roster.AddMember(writeCtx(t), candidate, adminActor, slog.New(slog.DiscardHandler))
	if !errors.Is(err, errRosterRecordInvalid) {
		t.Errorf("AddMember(zero-width address variant) error = %v, want it to wrap errRosterRecordInvalid", err)
	}
	if !errors.Is(err, ErrInvalidMemberAddress) {
		t.Errorf("AddMember(zero-width address variant) error = %v, want it to wrap ErrInvalidMemberAddress", err)
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
	actors := []Actor{adminActor, operator}
	for i, by := range actors {
		roster, _ := newRosterOver(t, liveAdmin)
		plain := Member{Address: fmt.Sprintf("plain%d@example.com", i), DisplayName: "A", Username: fmt.Sprintf("plain%d", i)}
		if err := roster.AddMember(writeCtx(t), plain, by, logger); err != nil {
			t.Fatalf("AddMember(plain fixture, %s) = %v, want nil", by.channel, err)
		}
	}

	for _, tt := range defaultIgnorableRoster {
		for _, by := range actors {
			t.Run(tt.name+"/"+by.channel.String(), func(t *testing.T) {
				roster, h := newRosterOver(t, liveAdmin)
				before := dumpUsers(t, h)

				candidate := Member{Address: defaultIgnorableAddress(tt.r), DisplayName: "A", Username: "a"}
				err := roster.AddMember(writeCtx(t), candidate, by, logger)
				if err == nil {
					t.Fatalf("AddMember() error = nil, want a refusal")
				}
				if !errors.Is(err, errRosterRecordInvalid) {
					t.Errorf("AddMember() error = %v, want it to wrap errRosterRecordInvalid", err)
				}
				if !errors.Is(err, ErrInvalidMemberAddress) {
					t.Errorf("AddMember() error = %v, want it to wrap ErrInvalidMemberAddress", err)
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
// logger's own, a member actor named by its id's wire form, and the
// command-line client address read as "local".
func TestAddMember_AuditRecord(t *testing.T) {
	alice := Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
	tests := []struct {
		name      string
		candidate Member
		by        Actor
		want      map[string]any
	}{
		{
			name:      "a cookie add",
			candidate: alice,
			by:        adminActor,
			want: map[string]any{
				"actor": adminWire, "channel": "cookie", "address": "alice@example.com",
				"admin": false, "client_ip": "198.51.100.7",
			},
		},
		{
			name:      "a command-line add",
			candidate: alice,
			by:        operator,
			want: map[string]any{
				"actor": "operator", "channel": "command-line", "address": "alice@example.com",
				"admin": false, "client_ip": "local",
			},
		},
		{
			name:      "a command-line admin add",
			candidate: Member{Address: "boss@example.com", DisplayName: "Boss", Username: "boss", Admin: true},
			by:        operator,
			want: map[string]any{
				"actor": "operator", "channel": "command-line", "address": "boss@example.com",
				"admin": true, "client_ip": "local",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roster, _ := newRosterOver(t, liveAdmin)
			var logBuf bytes.Buffer
			if err := roster.AddMember(writeCtx(t), tt.candidate, tt.by, slog.New(slog.NewJSONHandler(&logBuf, nil))); err != nil {
				t.Fatalf("AddMember: %v", err)
			}
			assertAuditRecord(t, logBuf.Bytes(), "roster member added", tt.want)
		})
	}
}

// assertAuditRecord fails t unless logged holds exactly one record, with
// message msg, carrying want's fields and nothing beside them but the
// logger's own.
func assertAuditRecord(t *testing.T, logged []byte, msg string, want map[string]any) {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logged), &record); err != nil {
		t.Fatalf("unmarshal one audit record from %q: %v", logged, err)
	}
	if record["msg"] != msg {
		t.Errorf("audit record msg = %v, want %q", record["msg"], msg)
	}
	for key, wantValue := range want {
		if gotValue, ok := record[key]; !ok || gotValue != wantValue {
			t.Errorf("audit record[%q] = %v (present %v), want %v", key, gotValue, ok, wantValue)
		}
	}
	for key, value := range record {
		if _, ok := want[key]; !ok && key != "time" && key != "level" && key != "msg" {
			t.Errorf("audit record carries unexpected key %q (value %v)", key, value)
		}
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
			errs[i] = roster.AddMember(writeCtx(t), candidate, operator, slog.New(slog.DiscardHandler))
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
// other refused by ErrDuplicateUsername. The collision read runs
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
		case errors.Is(err, ErrDuplicateUsername):
			refusals++
		default:
			t.Errorf("AddMember() error = %v, want nil or ErrDuplicateUsername", err)
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
			if tt.wantErr != errors.Is(err, ErrBlankDisplayName) {
				t.Errorf("validateWriteCandidate(display name %q) error = %v, want ErrBlankDisplayName %v", tt.displayName, err, tt.wantErr)
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
	return newUsersStoreAt(t, t.TempDir())
}

// newUsersStoreAt creates a current store under stateDir and returns its
// serve-mode handle, closed at cleanup.
func newUsersStoreAt(t *testing.T, stateDir string) *store.Handle {
	t.Helper()
	h, err := store.OpenServe(context.Background(), stateDir, store.ServeOptions{
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

// adminActor is liveAdmin acting over the cookie channel, and adminWire
// the audit record's name for it, written out as a literal.
var (
	adminActor = MemberActor(adminID, "198.51.100.7")
	adminWire  = "member:ad010000000000000000000000000000"
)

// operator is the command-line actor every operator fixture names.
var operator = OperatorActor("operator")

// rosterNow is the instant a fixture roster's clock reads unless the
// test moves it.
var rosterNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// fixedNow returns a clock that always reads at.
func fixedNow(at time.Time) func() time.Time {
	return func() time.Time { return at }
}

// testClock is a clock a test moves by hand while the roster reads it.
type testClock struct {
	mu sync.Mutex
	at time.Time
}

func (c *testClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

func (c *testClock) set(at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = at
}

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
			got, err := NewRoster(h, nil, nil).Lookup(t.Context(), tt.address)
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
		roster := NewRoster(h, nil, nil)
		if err := h.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
		_, err := roster.Lookup(t.Context(), "alice@example.com")
		if err == nil || errors.Is(err, ErrMemberNotFound) {
			t.Fatalf("Lookup on a closed store error = %v, want the store's error, not ErrMemberNotFound", err)
		}
	})

	t.Run("a stored id that is not 16 bytes", func(t *testing.T) {
		h := newUsersStore(t)
		insertShortIDUser(t, h, "short@example.com")
		got, err := NewRoster(h, nil, nil).Lookup(t.Context(), "short@example.com")
		if err == nil || errors.Is(err, ErrMemberNotFound) {
			t.Fatalf("Lookup of a short-id row = %+v, %v; want an error that is not ErrMemberNotFound", got, err)
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
		got, err := NewRoster(newUsersStore(t), nil, nil).List(t.Context())
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
		got, err := NewRoster(h, nil, nil).List(t.Context())
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
		got, err := NewRoster(h, nil, nil).List(t.Context())
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
		roster := NewRoster(h, nil, nil)
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
		got, err := NewRoster(h, nil, nil).List(t.Context())
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
// username row answers ErrDuplicateUsername only when the collision
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
		by        Actor
		wantErr   error
		wantText  string
		check     func(t *testing.T, h *store.Handle)
	}{
		{
			name:      "a new address and username",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "New@Example.com", DisplayName: "New", Username: "newbie"},
			by:        adminActor,
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
			by:        adminActor,
			wantErr:   ErrDuplicateAddress,
		},
		{
			name:      "a tombstoned row on the address, re-added with its own old username",
			rows:      []userRow{liveAdmin, gone},
			candidate: Member{Address: "gone@example.com", DisplayName: "New Name", Username: "gone"},
			by:        adminActor,
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
			by:        adminActor,
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
			by:        adminActor,
			wantErr:   ErrDuplicateUsername,
		},
		{
			name:      "a new address asking for a tombstoned row's username, differing in ASCII case",
			rows:      []userRow{liveAdmin, gone},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "GONE"},
			by:        adminActor,
			wantErr:   ErrDuplicateUsername,
		},
		{
			name:      "no username",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New"},
			by:        adminActor,
			wantErr:   ErrMissingUsername,
		},
		{
			name:      "admin requested on the cookie channel",
			rows:      []userRow{liveAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie", Admin: true},
			by:        adminActor,
			wantErr:   errAdminGrantForbidden,
		},
		{
			name:      "admin on the command-line channel",
			candidate: Member{Address: "first@example.com", DisplayName: "First", Username: "first", Admin: true},
			by:        operator,
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
			by:        MemberActor(identity.UserID{0x57, 0x4a}, "192.0.2.1"),
			wantErr:   ErrActorNotAdmin,
		},
		{
			name:      "a cookie actor with a non-admin live row",
			rows:      []userRow{liveAdmin, plain},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			by:        MemberActor(plainID, "192.0.2.1"),
			wantErr:   ErrActorNotAdmin,
		},
		{
			name:      "a cookie actor with only a tombstoned admin row",
			rows:      []userRow{liveAdmin, formerAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			by:        MemberActor(formerID, "192.0.2.1"),
			wantErr:   ErrActorNotAdmin,
		},
		{
			name:      "an admin with no username",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Admin: true},
			by:        operator,
			wantErr:   ErrMissingUsername,
		},
		{
			name:      "an admin with username .x",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: ".x", Admin: true},
			by:        operator,
			wantErr:   ErrInvalidUsername,
		},
		{
			name:      "an admin whose display name carries U+202E",
			candidate: Member{Address: "new@example.com", DisplayName: "Ne\u202ew", Username: "new", Admin: true},
			by:        operator,
			wantErr:   ErrInvalidDisplayName,
		},
		{
			name:      "an empty command-line actor",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "new", Admin: true},
			by:        OperatorActor(""),
			wantErr:   errInvalidActor,
		},
		{
			name:      "a command-line non-admin add into an empty table",
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			by:        operator,
			wantErr:   ErrNoAdmin,
			wantText:  "member add --admin",
		},
		{
			name:      "a command-line non-admin add beside only a tombstoned admin",
			rows:      []userRow{formerAdmin},
			candidate: Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"},
			by:        operator,
			wantErr:   ErrNoAdmin,
			wantText:  "member add --admin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUsersStore(t)
			insertUsers(t, h, tt.rows...)
			before := dumpUsers(t, h)

			var logBuf bytes.Buffer
			roster := NewRoster(h, newEpochStore(t, t.TempDir()), fixedNow(rosterNow))
			err := roster.AddMember(writeCtx(t), tt.candidate, tt.by, slog.New(slog.NewJSONHandler(&logBuf, nil)))
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

// rosterFixture is a revoking roster over a fresh store, with the
// pieces a removal test reads back: the store's handle and directory,
// the auth leaf's store and epoch store, and the clock the roster reads.
type rosterFixture struct {
	roster   *Roster
	h        *store.Handle
	storeDir string
	auth     *Store
	epoch    *EpochStore
	clock    *testClock
}

// newRosterFixture returns a rosterFixture whose store holds rows and
// whose clock reads rosterNow.
func newRosterFixture(t *testing.T, rows ...userRow) rosterFixture {
	t.Helper()
	fx := rosterFixture{storeDir: t.TempDir(), clock: &testClock{at: rosterNow}}
	fx.h = newUsersStoreAt(t, fx.storeDir)
	insertUsers(t, fx.h, rows...)
	authStore, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	fx.auth = authStore
	fx.epoch = OpenEpochStore(authStore)
	fx.roster = NewRoster(fx.h, fx.epoch, fx.clock.now)
	return fx
}

// epochOf returns address's stored epoch, and whether there is one.
func (fx rosterFixture) epochOf(t *testing.T, address string) (time.Time, bool) {
	t.Helper()
	epoch, ok, err := fx.epoch.Read(address)
	if err != nil {
		t.Fatalf("Read epoch: %v", err)
	}
	return epoch, ok
}

// freezeEpochs makes the epoch subdirectory unwritable, so every bump
// fails with an error that is not ErrEpochBackward, and proves it did.
func (fx rosterFixture) freezeEpochs(t *testing.T) {
	t.Helper()
	dir := fx.auth.subdir(epochSubdir)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("chmod the epoch subdirectory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, dirPerm) })
	probe, err := os.CreateTemp(dir, "probe")
	if err == nil {
		_ = probe.Close()
		t.Fatal("the epoch subdirectory still takes a file; the fixture cannot make a bump fail")
	}
}

// secondHandle opens another write handle on the fixture's store: the
// second connection a concurrency fixture holds the write lock from.
func (fx rosterFixture) secondHandle(t *testing.T) *store.Handle {
	t.Helper()
	h, err := store.OpenWrite(writeCtx(t), fx.storeDir, usersBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// lockHold is how long a concurrency fixture holds the write lock before
// committing: well inside usersBusy, so correct code never answers busy.
const lockHold = 300 * time.Millisecond

// whileHeld runs stmt on other inside an Update it holds open, starts
// write once the write lock is held, commits stmt after lockHold, and
// returns write's error. It fails t if the held Update does not commit.
func whileHeld(t *testing.T, other *store.Handle, stmt string, args []any, write func() error) error {
	t.Helper()
	locked := make(chan struct{})
	release := make(chan struct{})
	held := make(chan error, 1)
	go func() {
		held <- other.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
			if _, err := tx.Exec(ctx, stmt, args...); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	select {
	case <-locked:
	case err := <-held:
		t.Fatalf("the held write ended before taking the lock: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- write() }()
	time.Sleep(lockHold)
	close(release)
	if err := <-held; err != nil {
		t.Fatalf("the held write: %v", err)
	}
	return <-done
}

// insertMemberRecords writes one member_preferences, member_overrides,
// and member_quota row for id.
func insertMemberRecords(t *testing.T, h *store.Handle, id identity.UserID) {
	t.Helper()
	err := h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
		for _, stmt := range []string{
			`INSERT INTO member_preferences (user_id, collapse_duplicates, updated_at) VALUES (?, 1, 100)`,
			`INSERT INTO member_overrides (user_id, quota_bytes, updated_at) VALUES (?, 1000, 100)`,
			`INSERT INTO member_quota (user_id, charged_bytes, updated_at) VALUES (?, 500, 100)`,
		} {
			if _, err := tx.Exec(ctx, stmt, id[:]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert member records: %v", err)
	}
}

// memberRecords counts id's rows in each of the three per-member tables,
// in the order preferences, overrides, quota.
func memberRecords(t *testing.T, h *store.Handle, id identity.UserID) [3]int {
	t.Helper()
	var counts [3]int
	for i, table := range []string{"member_preferences", "member_overrides", "member_quota"} {
		if err := h.QueryRow(t.Context(), `SELECT count(*) FROM `+table+` WHERE user_id = ?`, id[:]).Scan(&counts[i]); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
	}
	return counts
}

// liveAdmins returns the ids of every live admin, in id order, as hex.
func liveAdmins(t *testing.T, h *store.Handle) []string {
	t.Helper()
	rows, err := h.Query(t.Context(), `SELECT lower(hex(id)) FROM users WHERE admin = 1 AND tombstoned_at IS NULL ORDER BY id`)
	if err != nil {
		t.Fatalf("read admins: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan admin: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read admins: %v", err)
	}
	return ids
}

// verifyAt verifies a session minted for address at minted, checked at
// minted plus a minute, through a seam over fx's roster and epoch store.
func (fx rosterFixture) verifyAt(t *testing.T, address string, minted time.Time) (VerifiedIdentity, error) {
	t.Helper()
	value, _, err := MintSession([]byte(currentFixture), address, minted, seamLifetime, seamCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	seam := NewMagicLinkSeam(fx.roster, fx.epoch, []byte(currentFixture), []byte(previousFixture), fixed(seamLifetime), fixed(seamCap))
	return seam.VerifyIdentity(requestWithSession(t, value), minted.Add(time.Minute), false)
}

// aliceRow is a live non-admin member every removal fixture removes.
var aliceRow = userRow{id: aliceID, address: "alice@example.com", displayName: "Alice", username: "alice", createdAt: 110}

// TestRemoveMember_Checklist is the checklist test for the 14a rows of
// the removal checklist (AC15): the removal tombstones the row with the
// removal time, deletes the member's preferences, keeps their overrides
// and quota, and bumps the address's epoch to the removal time. A
// session minted before the removal stays refused after the address is
// re-added under its old id, even with the removal's epoch record
// deleted first, because the re-add bumps too; a sign-in one second
// past the re-add's bump verifies. Fixture state: a preference, an
// override, and a quota row, written before the removal.
func TestRemoveMember_Checklist(t *testing.T) {
	fx := newRosterFixture(t, liveAdmin, aliceRow)
	insertMemberRecords(t, fx.h, aliceID)
	minted := rosterNow.Add(-time.Hour)
	if _, err := fx.verifyAt(t, "alice@example.com", minted); err != nil {
		t.Fatalf("the pre-removal session does not verify before the removal: %v", err)
	}

	removedAt := rosterNow
	if err := fx.roster.RemoveMember(writeCtx(t), aliceID, adminActor, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}

	var tombstonedAt int64
	if err := fx.h.QueryRow(t.Context(), `SELECT tombstoned_at FROM users WHERE id = ?`, aliceID[:]).Scan(&tombstonedAt); err != nil {
		t.Fatalf("read the tombstone: %v", err)
	}
	if want := removedAt.UnixMicro(); tombstonedAt != want {
		t.Errorf("tombstoned_at = %d, want the removal time %d", tombstonedAt, want)
	}
	if got, want := memberRecords(t, fx.h, aliceID), [3]int{0, 1, 1}; got != want {
		t.Errorf("preferences, overrides, quota after removal = %v, want %v", got, want)
	}
	if epoch, ok := fx.epochOf(t, "alice@example.com"); !ok || !epoch.Equal(removedAt) {
		t.Errorf("epoch after removal = %v (present %v), want the removal time %v", epoch, ok, removedAt)
	}
	if _, err := fx.verifyAt(t, "alice@example.com", minted); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("pre-removal session after removal: %v, want ErrSessionRevoked", err)
	}

	path, err := fx.auth.EpochPath(RecordID(epochSubdir, "alice@example.com"))
	if err != nil {
		t.Fatalf("EpochPath: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("delete the removal's epoch record: %v", err)
	}

	readdedAt := rosterNow.Add(time.Hour)
	fx.clock.set(readdedAt)
	readd := Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
	if err := fx.roster.AddMember(writeCtx(t), readd, adminActor, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	if u := readUser(t, fx.h, "alice@example.com"); !bytes.Equal(u.id, aliceID[:]) || u.tombstoned {
		t.Fatalf("re-added row = %+v, want the old id %x, live", u, aliceID[:])
	}
	if epoch, ok := fx.epochOf(t, "alice@example.com"); !ok || !epoch.Equal(readdedAt) {
		t.Errorf("epoch after re-add = %v (present %v), want the re-add time %v", epoch, ok, readdedAt)
	}
	if _, err := fx.verifyAt(t, "alice@example.com", minted); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("pre-removal session after re-add: %v, want ErrSessionRevoked", err)
	}
	if _, err := fx.verifyAt(t, "alice@example.com", readdedAt.Add(-time.Minute)); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("session minted between the removal and the re-add: %v, want ErrSessionRevoked", err)
	}

	got, err := fx.verifyAt(t, "alice@example.com", readdedAt.Add(time.Second))
	if err != nil {
		t.Fatalf("sign-in one second past the re-add's bump: %v, want it to verify", err)
	}
	if got.Identity.UserID != aliceID {
		t.Errorf("signed-in id = %x, want %x", got.Identity.UserID, aliceID)
	}
}

// TestRemoveMember_BumpFailureRollsBack proves a bump that fails with
// anything but ErrEpochBackward rolls the removal back: the member stays
// live with their preferences. Fixture state: an epoch subdirectory no
// bump can write.
func TestRemoveMember_BumpFailureRollsBack(t *testing.T) {
	fx := newRosterFixture(t, liveAdmin, aliceRow)
	insertMemberRecords(t, fx.h, aliceID)
	before := dumpUsers(t, fx.h)
	fx.freezeEpochs(t)

	var logBuf bytes.Buffer
	err := fx.roster.RemoveMember(writeCtx(t), aliceID, adminActor, slog.New(slog.NewJSONHandler(&logBuf, nil)))
	if err == nil || errors.Is(err, ErrEpochBackward) {
		t.Fatalf("RemoveMember with a failing bump = %v, want the bump's error", err)
	}
	if after := dumpUsers(t, fx.h); after != before {
		t.Errorf("a failed bump left the users table changed:\nbefore %s\nafter  %s", before, after)
	}
	if got, want := memberRecords(t, fx.h, aliceID), [3]int{1, 1, 1}; got != want {
		t.Errorf("preferences, overrides, quota after a failed removal = %v, want %v", got, want)
	}
	if logBuf.Len() != 0 {
		t.Errorf("a failed removal logged %q, want no audit record", logBuf.String())
	}
}

// TestAddMember_ReactivationBumpFailureRollsBack proves the re-add's bump
// runs under the removal's failure policy: a bump that fails with
// anything but ErrEpochBackward leaves the row tombstoned.
func TestAddMember_ReactivationBumpFailureRollsBack(t *testing.T) {
	fx := newRosterFixture(t, liveAdmin, userRow{id: goneID, address: "gone@example.com", displayName: "Gone", username: "gone", tombstoned: true, createdAt: 50})
	before := dumpUsers(t, fx.h)
	fx.freezeEpochs(t)

	readd := Member{Address: "gone@example.com", DisplayName: "Gone", Username: "gone"}
	err := fx.roster.AddMember(writeCtx(t), readd, adminActor, slog.New(slog.DiscardHandler))
	if err == nil || errors.Is(err, ErrEpochBackward) {
		t.Fatalf("re-add with a failing bump = %v, want the bump's error", err)
	}
	if after := dumpUsers(t, fx.h); after != before {
		t.Errorf("a failed re-add bump left the users table changed:\nbefore %s\nafter  %s", before, after)
	}
}

// TestRosterWrites_EpochBackwardCommits proves a bump refused with
// ErrEpochBackward, because the stored epoch is later than the clock,
// logs and commits, for a removal and for a re-add, and leaves the later
// epoch in place. Fixture state: the address's epoch bumped five hours
// past the roster's clock.
func TestRosterWrites_EpochBackwardCommits(t *testing.T) {
	later := rosterNow.Add(5 * time.Hour)
	tests := []struct {
		name  string
		write func(t *testing.T, fx rosterFixture, logger *slog.Logger) error
		check func(t *testing.T, fx rosterFixture)
	}{
		{
			name: "a removal",
			write: func(t *testing.T, fx rosterFixture, logger *slog.Logger) error {
				return fx.roster.RemoveMember(writeCtx(t), aliceID, adminActor, logger)
			},
			check: func(t *testing.T, fx rosterFixture) {
				if u := readUser(t, fx.h, "alice@example.com"); !u.tombstoned {
					t.Errorf("removed row = %+v, want tombstoned", u)
				}
			},
		},
		{
			name: "a re-add",
			write: func(t *testing.T, fx rosterFixture, logger *slog.Logger) error {
				if err := fx.h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
					_, err := tx.Exec(ctx, `UPDATE users SET tombstoned_at = 1 WHERE id = ?`, aliceID[:])
					return err
				}); err != nil {
					t.Fatalf("tombstone the fixture row: %v", err)
				}
				readd := Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
				return fx.roster.AddMember(writeCtx(t), readd, adminActor, logger)
			},
			check: func(t *testing.T, fx rosterFixture) {
				if u := readUser(t, fx.h, "alice@example.com"); u.tombstoned {
					t.Errorf("re-added row = %+v, want live", u)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newRosterFixture(t, liveAdmin, aliceRow)
			if err := fx.epoch.Bump("alice@example.com", later); err != nil {
				t.Fatalf("Bump the later epoch: %v", err)
			}
			var logBuf bytes.Buffer
			if err := tt.write(t, fx, slog.New(slog.NewJSONHandler(&logBuf, nil))); err != nil {
				t.Fatalf("write against a later epoch = %v, want nil", err)
			}
			tt.check(t, fx)
			if epoch, ok := fx.epochOf(t, "alice@example.com"); !ok || !epoch.Equal(later) {
				t.Errorf("epoch = %v (present %v), want the later one %v left in place", epoch, ok, later)
			}
			if !strings.Contains(logBuf.String(), `"level":"WARN"`) || !strings.Contains(logBuf.String(), "later than the clock") {
				t.Errorf("log = %q, want a warning naming the epoch later than the clock", logBuf.String())
			}
		})
	}
}

// TestRemoveDemote_NotFound proves a removal or a demotion of an unknown
// id, or of a removed member, answers ErrMemberNotFound and changes
// nothing, epoch included.
func TestRemoveDemote_NotFound(t *testing.T) {
	gone := userRow{id: goneID, address: "gone@example.com", displayName: "Gone", username: "gone", admin: true, tombstoned: true, createdAt: 50}
	writes := []struct {
		name  string
		write func(r *Roster, id identity.UserID) error
	}{
		{"remove", func(r *Roster, id identity.UserID) error {
			return r.RemoveMember(writeCtx(t), id, adminActor, slog.New(slog.DiscardHandler))
		}},
		{"demote", func(r *Roster, id identity.UserID) error {
			return r.DemoteMember(writeCtx(t), id, adminActor, slog.New(slog.DiscardHandler))
		}},
	}
	targets := []struct {
		name string
		id   identity.UserID
	}{
		{"an unknown id", identity.UserID{0x0b, 0x5e}},
		{"a removed member", goneID},
	}
	for _, w := range writes {
		for _, target := range targets {
			t.Run(w.name+"/"+target.name, func(t *testing.T) {
				fx := newRosterFixture(t, liveAdmin, aliceRow, gone)
				before := dumpUsers(t, fx.h)
				if err := w.write(fx.roster, target.id); !errors.Is(err, ErrMemberNotFound) {
					t.Fatalf("%s(%s) = %v, want ErrMemberNotFound", w.name, target.name, err)
				}
				if after := dumpUsers(t, fx.h); after != before {
					t.Errorf("a refused %s changed the users table:\nbefore %s\nafter  %s", w.name, before, after)
				}
				if _, ok := fx.epochOf(t, "gone@example.com"); ok {
					t.Errorf("a refused %s bumped an epoch", w.name)
				}
			})
		}
	}
}

// TestDemoteMember proves a demotion clears the flag and logs one record,
// a demotion of a non-admin writes and logs nothing and succeeds, and a
// demotion of the last admin answers ErrNoAdmin and changes nothing.
func TestDemoteMember(t *testing.T) {
	second := userRow{id: formerID, address: "second@example.com", displayName: "Second", username: "second", admin: true, createdAt: 120}

	t.Run("an admin beside another", func(t *testing.T) {
		fx := newRosterFixture(t, liveAdmin, second)
		var logBuf bytes.Buffer
		if err := fx.roster.DemoteMember(writeCtx(t), formerID, adminActor, slog.New(slog.NewJSONHandler(&logBuf, nil))); err != nil {
			t.Fatalf("DemoteMember: %v", err)
		}
		if u := readUser(t, fx.h, "second@example.com"); u.admin != 0 || u.tombstoned {
			t.Errorf("demoted row = %+v, want admin 0 and live", u)
		}
		assertAuditRecord(t, logBuf.Bytes(), "roster member demoted", map[string]any{
			"actor": adminWire, "channel": "cookie", "address": "second@example.com", "client_ip": "198.51.100.7",
		})
		if _, ok := fx.epochOf(t, "second@example.com"); ok {
			t.Error("a demotion bumped the member's epoch")
		}
	})

	t.Run("a non-admin", func(t *testing.T) {
		fx := newRosterFixture(t, liveAdmin, aliceRow)
		before := dumpUsers(t, fx.h)
		var logBuf bytes.Buffer
		if err := fx.roster.DemoteMember(writeCtx(t), aliceID, adminActor, slog.New(slog.NewJSONHandler(&logBuf, nil))); err != nil {
			t.Fatalf("DemoteMember(non-admin) = %v, want nil", err)
		}
		if after := dumpUsers(t, fx.h); after != before {
			t.Errorf("demoting a non-admin changed the users table:\nbefore %s\nafter  %s", before, after)
		}
		if logBuf.Len() != 0 {
			t.Errorf("demoting a non-admin logged %q, want nothing", logBuf.String())
		}
	})

	t.Run("the last admin, by the command line", func(t *testing.T) {
		fx := newRosterFixture(t, liveAdmin, aliceRow)
		before := dumpUsers(t, fx.h)
		if err := fx.roster.DemoteMember(writeCtx(t), adminID, operator, slog.New(slog.DiscardHandler)); !errors.Is(err, ErrNoAdmin) {
			t.Fatalf("DemoteMember(last admin) = %v, want ErrNoAdmin", err)
		}
		if after := dumpUsers(t, fx.h); after != before {
			t.Errorf("a refused demotion changed the users table:\nbefore %s\nafter  %s", before, after)
		}
	})
}

// TestRemoveMember_AuditRecord proves a removal logs one record on each
// channel, naming the actor, the channel, the address, and the client.
func TestRemoveMember_AuditRecord(t *testing.T) {
	tests := []struct {
		name string
		by   Actor
		want map[string]any
	}{
		{"a cookie removal", adminActor, map[string]any{
			"actor": adminWire, "channel": "cookie", "address": "alice@example.com", "client_ip": "198.51.100.7",
		}},
		{"a command-line removal", operator, map[string]any{
			"actor": "operator", "channel": "command-line", "address": "alice@example.com", "client_ip": "local",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newRosterFixture(t, liveAdmin, aliceRow)
			var logBuf bytes.Buffer
			if err := fx.roster.RemoveMember(writeCtx(t), aliceID, tt.by, slog.New(slog.NewJSONHandler(&logBuf, nil))); err != nil {
				t.Fatalf("RemoveMember: %v", err)
			}
			assertAuditRecord(t, logBuf.Bytes(), "roster member removed", tt.want)
		})
	}
}

// TestRemoveMember_SoleAdminSelfRemovalRefused proves the last-admin
// guard refuses a sole admin's self-removal before the bump, so the
// admin's session still verifies and no epoch is written.
func TestRemoveMember_SoleAdminSelfRemovalRefused(t *testing.T) {
	fx := newRosterFixture(t, liveAdmin, aliceRow)
	before := dumpUsers(t, fx.h)
	minted := rosterNow.Add(-time.Hour)

	if err := fx.roster.RemoveMember(writeCtx(t), adminID, adminActor, slog.New(slog.DiscardHandler)); !errors.Is(err, ErrNoAdmin) {
		t.Fatalf("sole admin's self-removal = %v, want ErrNoAdmin", err)
	}
	if after := dumpUsers(t, fx.h); after != before {
		t.Errorf("a refused self-removal changed the users table:\nbefore %s\nafter  %s", before, after)
	}
	if _, ok := fx.epochOf(t, "admin@example.com"); ok {
		t.Error("a refused self-removal bumped the admin's epoch")
	}
	if _, err := fx.verifyAt(t, "admin@example.com", minted); err != nil {
		t.Errorf("the sole admin's session after the refused self-removal: %v, want it to verify", err)
	}
}

// TestRemoveMember_SelfWithAnotherAdmin proves an admin may remove
// themself while another admin remains, and that the bump ends their own
// session.
func TestRemoveMember_SelfWithAnotherAdmin(t *testing.T) {
	second := userRow{id: formerID, address: "second@example.com", displayName: "Second", username: "second", admin: true, createdAt: 120}
	fx := newRosterFixture(t, liveAdmin, second)
	minted := rosterNow.Add(-time.Hour)

	if err := fx.roster.RemoveMember(writeCtx(t), adminID, adminActor, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("self-removal beside another admin: %v", err)
	}
	if _, err := fx.verifyAt(t, "admin@example.com", minted); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("the removed admin's session: %v, want ErrSessionRevoked", err)
	}
	if got, want := liveAdmins(t, fx.h), []string{"f0120000000000000000000000000000"}; !slices.Equal(got, want) {
		t.Errorf("live admins = %v, want %v", got, want)
	}
}

// TestDemoteMember_ConcurrentLastAdmins is AC16's concurrency proof: a
// second connection holds the write lock with one admin's demotion
// uncommitted while the other admin demotes themself; the first commits
// after lockHold. The guard reads inside the second demotion's own
// transaction, so it sees the first and refuses: exactly one admin
// remains. A guard read outside the transaction sees two admins and
// admits both.
func TestDemoteMember_ConcurrentLastAdmins(t *testing.T) {
	second := userRow{id: aliceID, address: "alice@example.com", displayName: "Alice", username: "alice", admin: true, createdAt: 110}
	fx := newRosterFixture(t, liveAdmin, second)
	other := fx.secondHandle(t)

	err := whileHeld(t, other, `UPDATE users SET admin = 0 WHERE id = ?`, []any{adminID[:]}, func() error {
		return fx.roster.DemoteMember(writeCtx(t), aliceID, MemberActor(aliceID, "198.51.100.7"), slog.New(slog.DiscardHandler))
	})
	if !errors.Is(err, ErrNoAdmin) {
		t.Errorf("the second demotion = %v, want ErrNoAdmin", err)
	}
	if got, want := liveAdmins(t, fx.h), []string{"a11ce000000000000000000000000000"}; !slices.Equal(got, want) {
		t.Errorf("live admins = %v, want exactly %v", got, want)
	}
}

// TestRemoveMember_ActorDemotedInFlight proves the actor's admin flag is
// read inside the removal's own transaction: a second connection holds
// the write lock with the actor's demotion uncommitted while the actor
// removes another member; once the demotion commits, the removal answers
// ErrActorNotAdmin and writes nothing. A flag read before the write
// would still see the actor as an admin.
func TestRemoveMember_ActorDemotedInFlight(t *testing.T) {
	second := userRow{id: formerID, address: "second@example.com", displayName: "Second", username: "second", admin: true, createdAt: 120}
	fx := newRosterFixture(t, liveAdmin, second, aliceRow)
	insertMemberRecords(t, fx.h, aliceID)
	other := fx.secondHandle(t)

	err := whileHeld(t, other, `UPDATE users SET admin = 0 WHERE id = ?`, []any{adminID[:]}, func() error {
		return fx.roster.RemoveMember(writeCtx(t), aliceID, adminActor, slog.New(slog.DiscardHandler))
	})
	if !errors.Is(err, ErrActorNotAdmin) {
		t.Fatalf("removal by an actor demoted in flight = %v, want ErrActorNotAdmin", err)
	}
	if u := readUser(t, fx.h, "alice@example.com"); u.tombstoned {
		t.Errorf("target row = %+v, want it still live", u)
	}
	if got, want := memberRecords(t, fx.h, aliceID), [3]int{1, 1, 1}; got != want {
		t.Errorf("preferences, overrides, quota = %v, want %v", got, want)
	}
	if _, ok := fx.epochOf(t, "alice@example.com"); ok {
		t.Error("a refused removal bumped the target's epoch")
	}
}

// TestRosterWrites_ActorDemotedBefore proves a removal, a demotion, and a
// cookie-channel add each refuse an actor whose row a direct store write
// demoted or tombstoned first, and change nothing.
func TestRosterWrites_ActorDemotedBefore(t *testing.T) {
	second := userRow{id: formerID, address: "second@example.com", displayName: "Second", username: "second", admin: true, createdAt: 120}
	writes := []struct {
		name  string
		write func(r *Roster) error
	}{
		{"remove", func(r *Roster) error {
			return r.RemoveMember(writeCtx(t), aliceID, adminActor, slog.New(slog.DiscardHandler))
		}},
		{"demote", func(r *Roster) error {
			return r.DemoteMember(writeCtx(t), formerID, adminActor, slog.New(slog.DiscardHandler))
		}},
		{"add", func(r *Roster) error {
			return r.AddMember(writeCtx(t), Member{Address: "new@example.com", DisplayName: "New", Username: "newbie"}, adminActor, slog.New(slog.DiscardHandler))
		}},
	}
	changes := []struct {
		name string
		stmt string
	}{
		{"demoted", `UPDATE users SET admin = 0 WHERE id = ?`},
		{"tombstoned", `UPDATE users SET tombstoned_at = 1 WHERE id = ?`},
	}
	for _, w := range writes {
		for _, change := range changes {
			t.Run(w.name+"/"+change.name, func(t *testing.T) {
				fx := newRosterFixture(t, liveAdmin, second, aliceRow)
				if err := fx.h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
					_, err := tx.Exec(ctx, change.stmt, adminID[:])
					return err
				}); err != nil {
					t.Fatalf("change the actor's row: %v", err)
				}
				before := dumpUsers(t, fx.h)
				if err := w.write(fx.roster); !errors.Is(err, ErrActorNotAdmin) {
					t.Fatalf("%s by a %s actor = %v, want ErrActorNotAdmin", w.name, change.name, err)
				}
				if after := dumpUsers(t, fx.h); after != before {
					t.Errorf("a refused %s changed the users table:\nbefore %s\nafter  %s", w.name, before, after)
				}
			})
		}
	}
}

// TestCheckAdmin proves the in-write admin check: nil for a live admin,
// ErrActorNotAdmin for a live non-admin, a removed admin, and an unknown
// id.
func TestCheckAdmin(t *testing.T) {
	formerAdmin := userRow{id: formerID, address: "former@example.com", displayName: "Former", username: "former", admin: true, tombstoned: true, createdAt: 60}
	tests := []struct {
		name    string
		id      identity.UserID
		wantErr error
	}{
		{"a live admin", adminID, nil},
		{"a live non-admin", aliceID, ErrActorNotAdmin},
		{"a removed admin", formerID, ErrActorNotAdmin},
		{"an unknown id", identity.UserID{0x0b, 0x5e}, ErrActorNotAdmin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newRosterFixture(t, liveAdmin, aliceRow, formerAdmin)
			var got error
			if err := fx.h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
				got = CheckAdmin(ctx, tx, tt.id)
				return nil
			}); err != nil {
				t.Fatalf("Update: %v", err)
			}
			if !errors.Is(got, tt.wantErr) || (tt.wantErr == nil && got != nil) {
				t.Errorf("CheckAdmin(%s) = %v, want %v", tt.name, got, tt.wantErr)
			}
		})
	}
}
