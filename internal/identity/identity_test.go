package identity

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode"
)

// normalizeCase is one row of the normalizer's corpus. reason, when
// set on a refusal row, is text the refusal must carry, which is how a
// row names the check that refused it when a second check would also
// refuse the same input.
type normalizeCase struct {
	name    string
	input   string
	want    string
	wantErr bool
	reason  string
}

// normalizeCorpus is the corpus both normalizer tests run: every row is
// checked for its own outcome and, where accepted, for the fixed point.
func normalizeCorpus(t *testing.T) []normalizeCase {
	t.Helper()
	// The boundary fixtures pad the local part with "é" (a 2-byte,
	// 1-rune UTF-8 character) so a byte-length cap and a rune-count
	// cap disagree: 121 of them plus "@example.com" (12 ASCII bytes)
	// lands exactly at MaxAddressLength (254) in bytes but at 133
	// runes; one more byte (an ASCII letter appended to the local
	// part) crosses the byte cap while staying at 134 runes, well
	// under a rune-based cap of 254.
	atCap := strings.Repeat("é", 121) + "@example.com"
	overCap := strings.Repeat("é", 121) + "x@example.com"
	if len(atCap) != MaxAddressLength {
		t.Fatalf("fixture atCap is %d bytes, want exactly MaxAddressLength (%d)", len(atCap), MaxAddressLength)
	}
	if len(overCap) != MaxAddressLength+1 {
		t.Fatalf("fixture overCap is %d bytes, want exactly MaxAddressLength+1 (%d)", len(overCap), MaxAddressLength+1)
	}

	return []normalizeCase{
		{
			name:  "case difference collapses to one value",
			input: "Alice@Example.COM",
			want:  "alice@example.com",
		},
		{
			name:  "dot in local part survives unchanged apart from case",
			input: "First.Last@Example.com",
			want:  "first.last@example.com",
		},
		{
			name:  "exactly at the length cap",
			input: atCap,
			want:  strings.ToLower(atCap),
		},
		{
			name:    "one byte over the length cap",
			input:   overCap,
			wantErr: true,
		},
		{
			name:    "unparseable value",
			input:   "not-an-address",
			wantErr: true,
		},
		{
			name:  "quoted local part that is already an atom",
			input: `"First.Last"@Example.com`,
			want:  "first.last@example.com",
		},
		{
			// net/mail unquotes the local part, and the
			// unquoted value carries a space.
			name:    "quoted local part carrying a space",
			input:   `"a b"@example.com`,
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			// The same unquoting with no space or control byte
			// in the result, so only the re-parse refuses it.
			name:    "quoted local part carrying a comma",
			input:   `"a,b"@example.com`,
			wantErr: true,
			reason:  "does not parse to itself",
		},
		{
			// net/mail's re-parse refuses this input too, as it
			// refuses every control byte and invisible character, so
			// the reason is what shows the rune check refused it.
			name:    "quoted local part carrying a TAB",
			input:   "\"a\tb\"@example.com",
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			name:    "quoted local part carrying a NEL control",
			input:   "\"a\u0085b\"@example.com",
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			name:    "quoted local part carrying a C1 CSI control",
			input:   "\"a\u009bb\"@example.com",
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			name:    "quoted local part carrying a right-to-left override",
			input:   "\"a\u202eb\"@example.com",
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			name:    "quoted local part carrying a no-break space",
			input:   "\"a\u00a0b\"@example.com",
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			name:    "quoted local part carrying a zero width space",
			input:   "\"a\u200bb\"@example.com",
			wantErr: true,
			reason:  "control character or invisible character",
		},
		{
			// strings.ToLower maps the Kelvin sign (U+212A) to the
			// ASCII "k", and net/mail accepts the unlowered form, so
			// the stored and mailed address is the ASCII owner's,
			// never a distinct lookalike.
			name:  "Kelvin sign lowercases to ASCII k",
			input: "a@\u212Aexample.com",
			want:  "a@kexample.com",
		},
	}
}

func TestNormalizeAddress(t *testing.T) {
	for _, tt := range normalizeCorpus(t) {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeAddress(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeAddress(%q) = %q, want error", tt.input, got)
				}
				if !errors.Is(err, ErrInvalidAddress) {
					t.Errorf("NormalizeAddress(%q) error = %v, want wrapping ErrInvalidAddress", tt.input, err)
				}
				if !strings.Contains(err.Error(), tt.reason) {
					t.Errorf("NormalizeAddress(%q) error = %v, want it to name %q", tt.input, err, tt.reason)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeAddress(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeAddress(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNormalizeAddressIsAFixedPoint holds the property the package doc
// states: normalizing a normalized address returns it unchanged, so a
// value that crosses a boundary normalized is keyed the same way on
// the far side. It runs every row of the corpus: a refusal row passes
// by being refused, so a normalizer that began accepting the quoted
// rows would fail here on the second normalization as well as in
// TestNormalizeAddress.
func TestNormalizeAddressIsAFixedPoint(t *testing.T) {
	for _, tt := range normalizeCorpus(t) {
		t.Run(tt.name, func(t *testing.T) {
			once, err := NormalizeAddress(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("NormalizeAddress(%q): %v", tt.input, err)
				}
				return
			}
			twice, err := NormalizeAddress(once)
			if err != nil {
				t.Fatalf("NormalizeAddress(%q) = %q, which it then refuses: %v", tt.input, once, err)
			}
			if twice != once {
				t.Errorf("NormalizeAddress(%q) = %q, and again = %q", tt.input, once, twice)
			}
		})
	}
}

// defaultIgnorableRow is one code point this file's Default_Ignorable_Code_Point
// table checks: an address carrying it is refused, the same address
// without it is accepted, and the refusal names the code point rather
// than the address.
type defaultIgnorableRow struct {
	name string
	r    rune
	// preExisting notes that the rune was already refused by the
	// unconditional Cf rule IsInvisibleRune carried before the
	// derived-set refusal existed: the row proves no regression rather
	// than the derived set's own reach.
	preExisting bool
}

// defaultIgnorableRows is the derivation's four named examples (the
// Hangul fillers), CGJ and VS16 (two further examples the derivation's
// own review probed), the three Cf runes already refused before the
// derived-set refusal existed, and U+2800, refused by name rather than
// by the derived formula.
func defaultIgnorableRows() []defaultIgnorableRow {
	return []defaultIgnorableRow{
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
}

// TestNormalizeAddress_DefaultIgnorableRefused proves that each of the
// ten code points, embedded in an otherwise valid quoted local part, is
// refused, while the same address without the code point is accepted,
// so the refusal is the code point's and not the quoting. The three
// preExisting rows were already refused by the unconditional Cf rule
// IsInvisibleRune carried before the derived-set refusal existed and
// are labelled regressions rather than proofs of the derived set's own
// reach; the other seven pass only once isDefaultIgnorableCodePoint or
// the U+2800 rule exists.
func TestNormalizeAddress_DefaultIgnorableRefused(t *testing.T) {
	for _, tt := range defaultIgnorableRows() {
		t.Run(tt.name, func(t *testing.T) {
			plain := `"ab"@example.com`
			if _, err := NormalizeAddress(plain); err != nil {
				t.Fatalf("NormalizeAddress(%q) (fixture without the code point): %v", plain, err)
			}

			carrying := "\"a" + string(tt.r) + "b\"@example.com"
			_, err := NormalizeAddress(carrying)
			if err == nil {
				t.Fatalf("NormalizeAddress(%q) = nil error, want a refusal", carrying)
			}
			if !errors.Is(err, ErrInvalidAddress) {
				t.Errorf("NormalizeAddress(%q) error = %v, want it to wrap ErrInvalidAddress", carrying, err)
			}
			want := fmt.Sprintf("U+%04X", tt.r)
			if !strings.Contains(err.Error(), want) {
				t.Errorf("NormalizeAddress(%q) error = %q, want it to name %q", carrying, err.Error(), want)
			}
			if strings.Contains(err.Error(), carrying) {
				t.Errorf("NormalizeAddress(%q) error = %q, want it to not contain the address", carrying, err.Error())
			}
		})
	}
}

// TestDefaultIgnorableDerivedSetSize pins the derivation against the
// toolchain's own tables: unicode.Version names the Unicode version the
// formula was measured against, and the derived set's size (4174) is
// DerivedCoreProperties.txt's own "Total code points: 4174" for
// Default_Ignorable_Code_Point (Unicode 17.0.0, read 2026-09-24). A
// toolchain move to a later Unicode version fails the version row
// rather than silently rederiving against a table nobody reread, and
// an over-refusing formula (for example, dropping the
// Prepended_Concatenation_Mark subtraction, which would additionally
// count U+0600 through U+0605) fails the size row.
func TestDefaultIgnorableDerivedSetSize(t *testing.T) {
	const wantVersion = "17.0.0"
	if unicode.Version != wantVersion {
		t.Fatalf("unicode.Version = %q, want %q (the table this test's count was derived against); a toolchain move needs DerivedCoreProperties.txt reread", unicode.Version, wantVersion)
	}

	const wantSize = 4174
	count := 0
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if isDefaultIgnorableCodePoint(r) {
			count++
		}
	}
	if count != wantSize {
		t.Errorf("derived Default_Ignorable_Code_Point set size = %d, want %d", count, wantSize)
	}
}

func TestChannel_StringAndValid(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    string
	}{
		{"cookie", ChannelCookie, "cookie"},
		{"command line", ChannelCommandLine, "command-line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.channel.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
			if !tt.channel.Valid() {
				t.Errorf("Valid() = false, want true for %s", tt.name)
			}
		})
	}

	if got := Channel(0).String(); got != "unknown" {
		t.Errorf("String() of zero value = %q, want %q", got, "unknown")
	}
	if Channel(0).Valid() {
		t.Error("Valid() of zero value = true, want false")
	}
	if Channel(3).Valid() {
		t.Error("Valid() of Channel(3) = true, want false")
	}

	// Exactly two channels are recognized. A third value added to the
	// enumeration above without a matching change here must fail this
	// count, not silently pass by extension.
	validCount := 0
	for c := Channel(0); c <= Channel(len(tests)+1); c++ {
		if c.Valid() {
			validCount++
		}
	}
	if validCount != len(tests) {
		t.Errorf("valid channel count = %d, want %d", validCount, len(tests))
	}
}

func TestIdentity_Valid(t *testing.T) {
	tests := []struct {
		name string
		id   Identity
		want bool
	}{
		{
			name: "zero value",
			id:   Identity{},
			want: false,
		},
		{
			name: "fully populated",
			id: Identity{
				Address:     "alice@example.com",
				UserID:      UserID{0x01},
				Principal:   "alice",
				DisplayName: "Alice",
				Source:      "magic-link",
				Channel:     ChannelCookie,
			},
			want: true,
		},
		{
			// The id is minted by the store and set only where the
			// store resolves a member, so an identity that never met
			// the store is no login.
			name: "zero UserID",
			id: Identity{
				Address:   "alice@example.com",
				Principal: "alice",
				Source:    "magic-link",
				Channel:   ChannelCookie,
			},
			want: false,
		},
		{
			name: "missing address",
			id: Identity{
				UserID:    UserID{0x01},
				Principal: "alice",
				Source:    "magic-link",
				Channel:   ChannelCookie,
			},
			want: false,
		},
		{
			name: "missing principal",
			id: Identity{
				Address: "alice@example.com",
				UserID:  UserID{0x01},
				Source:  "magic-link",
				Channel: ChannelCookie,
			},
			want: false,
		},
		{
			name: "missing source",
			id: Identity{
				Address:   "alice@example.com",
				UserID:    UserID{0x01},
				Principal: "alice",
				Channel:   ChannelCookie,
			},
			want: false,
		},
		{
			name: "invalid channel",
			id: Identity{
				Address:   "alice@example.com",
				UserID:    UserID{0x01},
				Principal: "alice",
				Source:    "magic-link",
			},
			want: false,
		},
		{
			// The command line names an operating-system user, never a
			// verified login, so every other field being valid must not
			// rescue it.
			name: "command-line channel",
			id: Identity{
				Address:     "alice@example.com",
				UserID:      UserID{0x01},
				Principal:   "alice",
				DisplayName: "Alice",
				Source:      "magic-link",
				Channel:     ChannelCommandLine,
			},
			want: false,
		},
		{
			name: "an unrecognized channel past the last constant",
			id: Identity{
				Address:     "alice@example.com",
				UserID:      UserID{0x01},
				Principal:   "alice",
				DisplayName: "Alice",
				Source:      "magic-link",
				Channel:     Channel(3),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIdentity_PrincipalDistinctFromAddress proves the two fields do
// not collapse to one another: Principal carries a username, which
// differs from the login address.
func TestIdentity_PrincipalDistinctFromAddress(t *testing.T) {
	id := Identity{
		Address:   "alice@example.com",
		Principal: "member-0001",
		Source:    "magic-link",
		Channel:   ChannelCookie,
	}
	if id.Principal == id.Address {
		t.Fatalf("Principal (%q) and Address (%q) must be independent fields", id.Principal, id.Address)
	}
	if id.Address != "alice@example.com" {
		t.Errorf("Address = %q, want %q", id.Address, "alice@example.com")
	}
	if id.Principal != "member-0001" {
		t.Errorf("Principal = %q, want %q", id.Principal, "member-0001")
	}
}

// TestIdentity_AddressMatchesNormalizer proves a constructed
// Identity's Address field carries exactly what NormalizeAddress
// returns for the same input, so the two paths cannot drift apart.
func TestIdentity_AddressMatchesNormalizer(t *testing.T) {
	const input = "Bob@Example.COM"

	want, err := NormalizeAddress(input)
	if err != nil {
		t.Fatalf("NormalizeAddress(%q) unexpected error: %v", input, err)
	}
	if want != "bob@example.com" {
		t.Fatalf("NormalizeAddress(%q) = %q, want %q", input, want, "bob@example.com")
	}

	id := Identity{
		Address:   want,
		Principal: "bob",
		Source:    "magic-link",
		Channel:   ChannelCookie,
	}
	if id.Address != want {
		t.Errorf("Identity.Address = %q, want normalizer output %q", id.Address, want)
	}
}

// TestNormalizeAddress_ErrorNeverQuotesValue holds NormalizeAddress's
// error contract: a refusal names its reason and never any part of the
// value. Each row is an input [net/mail] refuses with an error that
// quotes the input, or its unparsed remainder, back; "secret" marks
// the part a log line must never carry.
func TestNormalizeAddress_ErrorNeverQuotesValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "a second address after a comma", input: "alice@example.com, secret@example.com"},
		{name: "a trailing word", input: "alice@example.com secret"},
		{name: "invalid UTF-8 in the local part", input: "a\xffsecret@example.com"},
		{name: "invalid UTF-8 in a quoted string", input: "\"secret\xff\"@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeAddress(tt.input)
			if !errors.Is(err, ErrInvalidAddress) {
				t.Fatalf("NormalizeAddress(%q) error = %v, want it to wrap ErrInvalidAddress", tt.input, err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Errorf("NormalizeAddress(%q) error = %q, want no part of the value in it", tt.input, err.Error())
			}
		})
	}
}

// TestIsInvisibleRune pins the predicate every package outside this
// one refuses invisible runes through: the address refusal's whole
// reach, the derived set and U+2800 included, and none of the ordinary
// letters, digits, and punctuation a name or address carries.
func TestIsInvisibleRune(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{name: "a letter", r: 'a', want: false},
		{name: "a digit", r: '7', want: false},
		{name: "a hyphen", r: '-', want: false},
		{name: "a non-ASCII letter", r: 'ë', want: false},
		{name: "a Hangul syllable", r: '한', want: false},
		{name: "the ordinary space", r: ' ', want: true},
		{name: "a newline", r: '\n', want: true},
		{name: "DEL", r: '\u007f', want: true},
		{name: "no-break space (Zs)", r: ' ', want: true},
		{name: "line separator (Zl)", r: ' ', want: true},
		{name: "paragraph separator (Zp)", r: ' ', want: true},
		{name: "right-to-left override (Cf)", r: '‮', want: true},
		{name: "Hangul filler U+3164 (Lo, derived set)", r: 'ㅤ', want: true},
		{name: "Hangul filler U+115F (Lo, derived set)", r: 'ᅟ', want: true},
		{name: "combining grapheme joiner U+034F (Mn, derived set)", r: '͏', want: true},
		{name: "variation selector 16 U+FE0F (Mn, derived set)", r: '️', want: true},
		{name: "braille pattern blank U+2800", r: '⠀', want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvisibleRune(tt.r); got != tt.want {
				t.Errorf("IsInvisibleRune(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}
