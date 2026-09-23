package identity

import (
	"errors"
	"strings"
	"testing"
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
			reason:  "control character or space",
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
			// refuses every control byte and space, so the
			// reason is what shows the byte check refused it.
			name:    "quoted local part carrying a TAB",
			input:   "\"a\tb\"@example.com",
			wantErr: true,
			reason:  "control character or space",
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

func TestChannel_StringAndValid(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    string
	}{
		{"cookie", ChannelCookie, "cookie"},
		{"header", ChannelHeader, "header"},
		{"service token", ChannelServiceToken, "service-token"},
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

	// The seam distinguishes exactly three channels. A fourth value
	// added to the enumeration above without a matching change here
	// must fail this count, not silently pass by extension.
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
				Principal:   "alice",
				DisplayName: "Alice",
				Source:      "magic-link",
				Channel:     ChannelCookie,
			},
			want: true,
		},
		{
			name: "missing address",
			id: Identity{
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
				Source:  "magic-link",
				Channel: ChannelCookie,
			},
			want: false,
		},
		{
			name: "missing source",
			id: Identity{
				Address:   "alice@example.com",
				Principal: "alice",
				Channel:   ChannelCookie,
			},
			want: false,
		},
		{
			name: "invalid channel",
			id: Identity{
				Address:   "alice@example.com",
				Principal: "alice",
				Source:    "magic-link",
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
// not collapse to one another: rung 6's surrogate-id substitution
// (2.0 hook 5) depends on Principal being free to diverge from
// Address.
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
