package musicbrainz

import (
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

func creditPart(name, joinPhrase, artistID, artistName string) CreditPart {
	return CreditPart{
		Name:       name,
		JoinPhrase: joinPhrase,
		Artist:     Artist{ID: artistID, Name: artistName, SortName: artistName},
	}
}

// TestParseArtistCreditsJoinPhrases is T8's measured trailing-joinphrase
// behavior. beets extends the parts list with [name, joinphrase] per
// element and joins the whole thing (musicbrainz.py:344-349), so the last
// element's joinphrase survives into the weight-3.0 album artist penalty.
// Dropping it from the join leaves the third row's expected trailing
// separator unproduced.
func TestParseArtistCreditsJoinPhrases(t *testing.T) {
	tests := []struct {
		name         string
		parts        []CreditPart
		wantArtist   string
		wantArtistID string
	}{
		{
			name: "an empty final joinphrase joins cleanly",
			parts: []CreditPart{
				creditPart("A", " & ", "artist-a", "A"),
				creditPart("B", "", "artist-b", "B"),
			},
			wantArtist:   "A & B",
			wantArtistID: "artist-a",
		},
		{
			name:         "a single credit keeps its trailing joinphrase",
			parts:        []CreditPart{creditPart("A", " feat. ", "artist-a", "A")},
			wantArtist:   "A feat. ",
			wantArtistID: "artist-a",
		},
		{
			name: "a repeated final joinphrase trails the last name",
			parts: []CreditPart{
				creditPart("A", " & ", "artist-a", "A"),
				creditPart("B", " & ", "artist-b", "B"),
			},
			wantArtist:   "A & B & ",
			wantArtistID: "artist-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseArtistCredits(tt.parts, matcher.DefaultTunables())
			if got.Artist != tt.wantArtist {
				t.Errorf("Artist = %q, want %q", got.Artist, tt.wantArtist)
			}
			if got.ArtistID != tt.wantArtistID {
				t.Errorf("ArtistID = %q, want %q", got.ArtistID, tt.wantArtistID)
			}
		})
	}
}

// TestParseArtistCreditsPrefersTheAliasName is T8's A/B table. The
// fixture's one credit element carries three distinct names: the credit
// element's own display name, the artist entity's name, and a ja primary
// alias name. Only the entity-or-alias name reaches Artist, because
// artist_object is `alias or el["artist"]` (musicbrainz.py:328-330) while
// the credit string, which this port does not produce, reads el itself.
//
// The credit string is the one column of the inventory's table with no Go
// expression here: matcher.AlbumInfo carries no artist_credit and the
// ruling-E ledger does not grade one, so credits has no field for it and
// aliases_as_credits is not a tunable at all (R8). The two rows of that
// table with aliases_as_credits true are therefore unportable, and this
// test covers what remains: a mutation reading el.Name into Artist breaks
// the first row, and one dropping the alias lookup breaks the second.
func TestParseArtistCreditsPrefersTheAliasName(t *testing.T) {
	parts := loadTestdata[[]CreditPart](t, "artist_credit_alias.json")

	tests := []struct {
		name      string
		languages []string
		want      string
	}{
		{name: "no configured locale takes the entity name", want: "RealName"},
		{name: "a configured locale takes the alias name", languages: []string{"ja"}, want: "AliasName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := matcher.DefaultTunables()
			tun.Languages = tt.languages

			got := parseArtistCredits(parts, tun)
			if got.Artist != tt.want {
				t.Errorf("Artist = %q, want %q", got.Artist, tt.want)
			}
			if got.ArtistID != "artist-real" {
				t.Errorf("ArtistID = %q, want %q: the alias never moves the id", got.ArtistID, "artist-real")
			}
		})
	}
}

// TestApplyVariousArtists covers both branches of musicbrainz.py:673-681
// and the mutation the plan names: keying the override on the artist name
// rather than on the entity id. The third row is that proof, an artist
// credited as the various-artists string under an ordinary id, which beets
// leaves alone.
func TestApplyVariousArtists(t *testing.T) {
	tests := []struct {
		name       string
		in         credits
		vaName     string
		wantArtist string
		wantVA     bool
	}{
		{
			name:       "the various-artists id fires the override",
			in:         credits{Artist: "Assorted Credits & ", ArtistID: variousArtistsID},
			vaName:     "Various Artists",
			wantArtist: "Various Artists",
			wantVA:     true,
		},
		{
			name:       "the override writes the configured name",
			in:         credits{Artist: "Assorted Credits & ", ArtistID: variousArtistsID},
			vaName:     "Sundry Performers",
			wantArtist: "Sundry Performers",
			wantVA:     true,
		},
		{
			name:       "the name alone does not fire it",
			in:         credits{Artist: "Various Artists", ArtistID: "artist-ordinary"},
			vaName:     "Sundry Performers",
			wantArtist: "Various Artists",
			wantVA:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := matcher.DefaultTunables()
			tun.VAName = tt.vaName

			got, va := applyVariousArtists(tt.in, tun)
			if got.Artist != tt.wantArtist {
				t.Errorf("Artist = %q, want %q", got.Artist, tt.wantArtist)
			}
			if va != tt.wantVA {
				t.Errorf("va = %v, want %v", va, tt.wantVA)
			}
			if got.ArtistID != tt.in.ArtistID {
				t.Errorf("ArtistID = %q, want it untouched at %q", got.ArtistID, tt.in.ArtistID)
			}
		})
	}
}

// TestVariousArtistsID pins the literal against pinned beets'
// VARIOUS_ARTISTS_ID (beetsplug/musicbrainz.py:43). A transcription slip
// in it would silently disable the override on every VA release, which
// the corpus would show only as a moved album artist penalty.
func TestVariousArtistsID(t *testing.T) {
	const want = "89ad4ac3-39f7-470e-963a-56509c546377"
	if variousArtistsID != want {
		t.Errorf("variousArtistsID = %q, want %q", variousArtistsID, want)
	}
}
