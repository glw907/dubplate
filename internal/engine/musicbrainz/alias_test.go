package musicbrainz

import (
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// TestPreferredAlias is T7's measured table, row for row, against one
// fixture alias list carrying two ja primaries of different types, an en
// primary, an uppercase-JA primary, and a ja alias whose primary flag is
// null. The expectations are pinned beets' own answers, not this port's.
//
// Three of the rows are the mutation proofs the plan names: lowercasing
// the locale comparison breaks the JA row, returning nil instead of
// falling through breaks the ignored-artist-name row, and transposing the
// two loops breaks the en-before-ja row, because response order would
// then beat configured order.
func TestPreferredAlias(t *testing.T) {
	aliases := loadTestdata[[]Alias](t, "preferred_aliases.json")

	tests := []struct {
		name              string
		languages         []string
		ignoredAliasTypes []string
		want              string // the selected alias name, "" for nil
	}{
		{name: "shipped empty default selects nothing", want: ""},
		{name: "ja takes the first ja primary", languages: []string{"ja"}, want: "JaFirst"},
		{
			name:              "an ignored type falls through to the next ja primary",
			languages:         []string{"ja"},
			ignoredAliasTypes: []string{"artist name"},
			want:              "JaSecond",
		},
		{
			name:              "the ignored set is lowercased",
			languages:         []string{"ja"},
			ignoredAliasTypes: []string{"ARTIST NAME"},
			want:              "JaSecond",
		},
		{name: "the locale comparison is case-sensitive", languages: []string{"JA"}, want: "JaUpperLocale"},
		{name: "configured order beats response order", languages: []string{"en", "ja"}, want: "EnAlias"},
		{name: "the same list in the other order takes ja", languages: []string{"ja", "en"}, want: "JaFirst"},
		{
			name:              "every candidate ignored selects nothing",
			languages:         []string{"ja"},
			ignoredAliasTypes: []string{"artist name", "search hint"},
			want:              "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := matcher.DefaultTunables()
			tun.Languages = tt.languages
			tun.IgnoredAliasTypes = tt.ignoredAliasTypes

			got := preferredAlias(aliases, tun)
			if tt.want == "" {
				if got != nil {
					t.Fatalf("preferredAlias() = %q, want nil", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("preferredAlias() = nil, want %q", tt.want)
			}
			if got.Name != tt.want {
				t.Errorf("preferredAlias() = %q, want %q", got.Name, tt.want)
			}
		})
	}
}

// TestPreferredAliasSkipsNullPrimary isolates the rule the eight-row
// table only implies: the wire value is bool | null and a null is falsy
// (_utils/musicbrainz.py:138-155), so the guard is a nil check and a
// dereference. The fixture's null-primary ja alias sits ahead of every
// other ja alias in response order, so a port that treated the pointer
// itself as the flag would select it.
func TestPreferredAliasSkipsNullPrimary(t *testing.T) {
	aliases := loadTestdata[[]Alias](t, "preferred_aliases.json")
	if aliases[0].Primary != nil {
		t.Fatalf("fixture alias 0 primary = %v, want a null primary in response-order position 0", *aliases[0].Primary)
	}

	tun := matcher.DefaultTunables()
	tun.Languages = []string{"ja"}

	got := preferredAlias(aliases, tun)
	if got == nil {
		t.Fatal("preferredAlias() = nil, want the first ja primary")
	}
	if got.Name == aliases[0].Name {
		t.Errorf("preferredAlias() = %q, want the null-primary alias skipped", got.Name)
	}
}

// TestPreferredAliasOnEmptyList covers the branch beets short-circuits at
// musicbrainz.py:130 and the deliberate divergence R10 records: a
// recording carrying no aliases key at all decodes to a nil slice here,
// where beets raises KeyError at musicbrainz.py:589.
func TestPreferredAliasOnEmptyList(t *testing.T) {
	tun := matcher.DefaultTunables()
	tun.Languages = []string{"ja"}

	if got := preferredAlias(nil, tun); got != nil {
		t.Errorf("preferredAlias(nil) = %q, want nil", got.Name)
	}
}

// TestKeyWithPreferredAlias asserts both arms of the fallback beets
// spells at musicbrainz.py:158-162. The two fallback rows matter
// because the track-title override reads the alias arm as a guard
// rather than for the name it returns (musicbrainz.py:589-590).
func TestKeyWithPreferredAlias(t *testing.T) {
	aliases := loadTestdata[[]Alias](t, "preferred_aliases.json")

	tests := []struct {
		name      string
		aliases   []Alias
		languages []string
		fallback  string
		want      string
	}{
		{name: "no alias selected keeps the fallback", aliases: aliases, fallback: "WireTitle", want: "WireTitle"},
		{name: "no aliases at all keeps the fallback", languages: []string{"ja"}, fallback: "WireTitle", want: "WireTitle"},
		{name: "a selected alias replaces the fallback", aliases: aliases, languages: []string{"ja"}, fallback: "WireTitle", want: "JaFirst"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := matcher.DefaultTunables()
			tun.Languages = tt.languages

			if got := keyWithPreferredAlias(tt.aliases, tt.fallback, tun); got != tt.want {
				t.Errorf("keyWithPreferredAlias() = %q, want %q", got, tt.want)
			}
		})
	}
}
