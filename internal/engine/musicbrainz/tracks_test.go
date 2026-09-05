package musicbrainz

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

func titlesOf(tracks []builtTrack) []string {
	titles := make([]string, len(tracks))
	for i, tr := range tracks {
		titles[i] = tr.Info.Title
	}
	return titles
}

func simpleCreditPart(id, name string) CreditPart {
	return CreditPart{Name: name, Artist: Artist{ID: id, Name: name}}
}

// TestTracksFromMedium_PregapSurvivesUnconditionally proves T5's
// assembly order: the pregap has no config gate at all, unlike the data
// tracks it sits beside. Mutation: moving the pregap behind
// IgnoreDataTracks makes the true row lose "Pregap".
func TestTracksFromMedium_PregapSurvivesUnconditionally(t *testing.T) {
	m := loadTestdata[Medium](t, "gap_pregap_medium.json")
	for _, ignoreDataTracks := range []bool{true, false} {
		tun := matcher.DefaultTunables()
		tun.IgnoreDataTracks = ignoreDataTracks
		got := titlesOf(tracksFromMedium(m, tun))
		want := []string{"Pregap", "A"}
		if !slices.Equal(got, want) {
			t.Errorf("IgnoreDataTracks=%v: titles = %v, want %v", ignoreDataTracks, got, want)
		}
	}
}

// TestTracksFromMedium_DataTrackTruthTable reproduces T5's measured
// four-row table over a medium carrying a pregap, one audio track, one
// video track, one data track, and one "[data track]"-titled entry: the
// filtered title never survives any combination, and IgnoreDataTracks
// gates only whether the data medium contributes at all.
func TestTracksFromMedium_DataTrackTruthTable(t *testing.T) {
	m := loadTestdata[Medium](t, "gap_data_tracks_medium.json")
	tests := []struct {
		ignoreDataTracks  bool
		ignoreVideoTracks bool
		want              []string
	}{
		{true, true, []string{"Pregap", "A"}},
		{false, true, []string{"Pregap", "A", "D"}},
		{true, false, []string{"Pregap", "A", "V"}},
		{false, false, []string{"Pregap", "A", "V", "D"}},
	}
	for _, tt := range tests {
		tun := matcher.DefaultTunables()
		tun.IgnoreDataTracks = tt.ignoreDataTracks
		tun.IgnoreVideoTracks = tt.ignoreVideoTracks
		got := titlesOf(tracksFromMedium(m, tun))
		if !slices.Equal(got, tt.want) {
			t.Errorf("IgnoreDataTracks=%v IgnoreVideoTracks=%v: titles = %v, want %v",
				tt.ignoreDataTracks, tt.ignoreVideoTracks, got, tt.want)
		}
	}
}

// TestTracksFromMedium_BracketTitleFilteredInsideTracks proves the
// "[data track]" filter runs regardless of where the track lives:
// beets drops it whether it came through medium.Tracks or
// medium.DataTracks, and regardless of IgnoreDataTracks.
func TestTracksFromMedium_BracketTitleFilteredInsideTracks(t *testing.T) {
	m := loadTestdata[Medium](t, "gap_bracket_title_in_tracks.json")
	tun := matcher.DefaultTunables()
	tun.IgnoreDataTracks = false
	got := titlesOf(tracksFromMedium(m, tun))
	want := []string{"A"}
	if !slices.Equal(got, want) {
		t.Errorf("titles = %v, want %v", got, want)
	}
}

// TestTracksFromMedium_LengthFallbackIsTruthiness proves T6's fallback
// tests truthiness, not nil-ness: a track length of 0 falls back to the
// recording length exactly as a null length does. Mutation: comparing
// tr.Length != nil instead of *tr.Length != 0 makes the "zero" row keep
// 0 instead of falling back to 90.0.
func TestTracksFromMedium_LengthFallbackIsTruthiness(t *testing.T) {
	recordingLength := 90000.0
	zero := 0.0
	nonZero := 45000.0

	tests := []struct {
		name        string
		trackLength *float64
		wantSeconds float64
	}{
		{"null falls back", nil, 90.0},
		{"zero falls back", &zero, 90.0},
		{"real value wins", &nonZero, 45.0},
	}
	for _, tt := range tests {
		m := Medium{
			Position: 1,
			Format:   new("CD"),
			Tracks: []Track{{
				ID:       "trk-1",
				Position: 1,
				Number:   "1",
				Title:    "A",
				Length:   tt.trackLength,
				Recording: Recording{
					ID:           "rec-1",
					Title:        "A",
					Length:       &recordingLength,
					ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
				},
			}},
		}
		got := tracksFromMedium(m, matcher.DefaultTunables())
		if len(got) != 1 {
			t.Fatalf("%s: got %d tracks, want 1", tt.name, len(got))
		}
		if got[0].Info.Length != tt.wantSeconds {
			t.Errorf("%s: Length = %v, want %v", tt.name, got[0].Info.Length, tt.wantSeconds)
		}
	}
}

// TestTracksFromMedium_LengthConvertsMillisecondsToSeconds asserts the
// wire's millisecond unit becomes TrackInfo.Length's seconds
// (hooks.py:481).
func TestTracksFromMedium_LengthConvertsMillisecondsToSeconds(t *testing.T) {
	lengthMS := 200000.0
	m := Medium{
		Position: 1,
		Format:   new("CD"),
		Tracks: []Track{{
			ID: "trk-1", Position: 1, Number: "1", Title: "A",
			Recording: Recording{
				ID: "rec-1", Title: "A", Length: &lengthMS,
				ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
			},
		}},
	}
	got := tracksFromMedium(m, matcher.DefaultTunables())
	if len(got) != 1 {
		t.Fatalf("got %d tracks, want 1", len(got))
	}
	if got[0].Info.Length != 200.0 {
		t.Errorf("Length = %v, want 200.0", got[0].Info.Length)
	}
}

// TestTracksFromMedium_ArtistCreditFallbackIsTruthiness proves an empty
// track-level artist-credit slice falls back to the recording's, the
// same truthiness rule the length override carries.
func TestTracksFromMedium_ArtistCreditFallbackIsTruthiness(t *testing.T) {
	tests := []struct {
		name        string
		trackCredit []CreditPart
		wantArtist  string
	}{
		{"nil falls back", nil, "Recording Artist"},
		{"empty falls back", []CreditPart{}, "Recording Artist"},
		{"non-empty wins", []CreditPart{simpleCreditPart("art-2", "Track Artist")}, "Track Artist"},
	}
	for _, tt := range tests {
		m := Medium{
			Position: 1,
			Format:   new("CD"),
			Tracks: []Track{{
				ID: "trk-1", Position: 1, Number: "1", Title: "A",
				ArtistCredit: tt.trackCredit,
				Recording: Recording{
					ID: "rec-1", Title: "A",
					ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Recording Artist")},
				},
			}},
		}
		got := tracksFromMedium(m, matcher.DefaultTunables())
		if len(got) != 1 {
			t.Fatalf("%s: got %d tracks, want 1", tt.name, len(got))
		}
		if got[0].Info.Artist != tt.wantArtist {
			t.Errorf("%s: Artist = %q, want %q", tt.name, got[0].Info.Artist, tt.wantArtist)
		}
	}
}

// TestTracksFromMedium_TitleOverrideGuardedByPreferredAlias reproduces
// T6's measured precedence table: the track title only overrides the
// recording title when preferredAlias finds nothing on the recording's
// aliases, and the final title always passes back through
// keyWithPreferredAlias, which is why an alias wins even when
// track.Title is truthy.
func TestTracksFromMedium_TitleOverrideGuardedByPreferredAlias(t *testing.T) {
	jaLocale := "ja"
	primary := true
	jaAlias := Alias{Name: "AliasTitle", Locale: &jaLocale, Primary: &primary}

	tests := []struct {
		name       string
		languages  []string
		aliases    []Alias
		trackTitle string
		want       string
	}{
		{"alias wins when languages configured", []string{"ja"}, []Alias{jaAlias}, "TrackTitle", "AliasTitle"},
		{"no languages, track title wins", nil, []Alias{jaAlias}, "TrackTitle", "TrackTitle"},
		{"no alias present, track title wins", []string{"ja"}, nil, "TrackTitle", "TrackTitle"},
		{"no alias, empty track title falls to recording", []string{"ja"}, nil, "", "RecordingTitle"},
	}
	for _, tt := range tests {
		m := Medium{
			Position: 1,
			Format:   new("CD"),
			Tracks: []Track{{
				ID: "trk-1", Position: 1, Number: "1", Title: tt.trackTitle,
				Recording: Recording{
					ID: "rec-1", Title: "RecordingTitle", Aliases: tt.aliases,
					ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
				},
			}},
		}
		tun := matcher.DefaultTunables()
		tun.Languages = tt.languages
		got := tracksFromMedium(m, tun)
		if len(got) != 1 {
			t.Fatalf("%s: got %d tracks, want 1", tt.name, len(got))
		}
		if got[0].Info.Title != tt.want {
			t.Errorf("%s: Title = %q, want %q", tt.name, got[0].Info.Title, tt.want)
		}
	}
}

// TestExtractTracks_MediumIndexDecodesBothWireSpellings proves the
// two-shape position decode carries all the way through to
// MediumIndex: a JSON string position and a JSON number position must
// both decode to the same int (fold, B3).
func TestExtractTracks_MediumIndexDecodesBothWireSpellings(t *testing.T) {
	raw := []byte(`{
		"media": [{
			"position": 1,
			"format": "CD",
			"tracks": [
				{"id": "trk-string", "position": "3", "number": "3", "title": "A",
				 "recording": {"id": "rec-1", "title": "A",
					"artist-credit": [{"name": "Artist", "artist": {"id": "art-1", "name": "Artist"}}]}},
				{"id": "trk-number", "position": 4, "number": "4", "title": "B",
				 "recording": {"id": "rec-2", "title": "B",
					"artist-credit": [{"name": "Artist", "artist": {"id": "art-1", "name": "Artist"}}]}}
			]
		}]
	}`)
	var rel Release
	if err := json.Unmarshal(raw, &rel); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	tracks := extractTracks(&rel, matcher.DefaultTunables())
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want 2", len(tracks))
	}
	if tracks[0].Info.MediumIndex != 3 {
		t.Errorf("string position: MediumIndex = %d, want 3", tracks[0].Info.MediumIndex)
	}
	if tracks[1].Info.MediumIndex != 4 {
		t.Errorf("number position: MediumIndex = %d, want 4", tracks[1].Info.MediumIndex)
	}
}

// TestExtractTracks_IndexAssignedAcrossAllMedia proves Index is
// 1-based across every surviving medium rather than restarting per
// medium (musicbrainz.py:645-646). Mutation: assigning Index inside
// tracksFromMedium instead of extractTracks makes the second medium's
// first track claim Index 1 instead of 3.
func TestExtractTracks_IndexAssignedAcrossAllMedia(t *testing.T) {
	rel := Release{Media: []Medium{
		{
			Position: 1, Format: new("CD"),
			Tracks: []Track{
				{ID: "t1", Position: 1, Number: "1", Title: "A", Recording: Recording{ID: "r1", Title: "A", ArtistCredit: []CreditPart{simpleCreditPart("a1", "Artist")}}},
				{ID: "t2", Position: 2, Number: "2", Title: "B", Recording: Recording{ID: "r2", Title: "B", ArtistCredit: []CreditPart{simpleCreditPart("a1", "Artist")}}},
			},
		},
		{
			Position: 2, Format: new("CD"),
			Tracks: []Track{
				{ID: "t3", Position: 1, Number: "1", Title: "C", Recording: Recording{ID: "r3", Title: "C", ArtistCredit: []CreditPart{simpleCreditPart("a1", "Artist")}}},
				{ID: "t4", Position: 2, Number: "2", Title: "D", Recording: Recording{ID: "r4", Title: "D", ArtistCredit: []CreditPart{simpleCreditPart("a1", "Artist")}}},
			},
		},
	}}
	tracks := extractTracks(&rel, matcher.DefaultTunables())
	if len(tracks) != 4 {
		t.Fatalf("got %d tracks, want 4", len(tracks))
	}
	wantIndices := []int{1, 2, 3, 4}
	for i, want := range wantIndices {
		if tracks[i].Info.Index != want {
			t.Errorf("track %d: Index = %d, want %d", i, tracks[i].Info.Index, want)
		}
	}
}

// TestExtractTracks_IgnoredMediaIsCaseSensitive proves IgnoredMedia
// membership is exact string comparison: a configured "cd" does not
// exclude a medium whose format is "CD".
func TestExtractTracks_IgnoredMediaIsCaseSensitive(t *testing.T) {
	rel := Release{Media: []Medium{{
		Position: 1, Format: new("CD"),
		Tracks: []Track{{ID: "t1", Position: 1, Number: "1", Title: "A", Recording: Recording{ID: "r1", Title: "A", ArtistCredit: []CreditPart{simpleCreditPart("a1", "Artist")}}}},
	}}}
	tun := matcher.DefaultTunables()
	tun.IgnoredMedia = []string{"cd"}
	tracks := extractTracks(&rel, tun)
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1 (the \"CD\" medium must survive an [\"cd\"] ignore list)", len(tracks))
	}
}

// TestExtractTracks_NullFormatMediumNeverIgnored proves a null format
// never matches any configured IgnoredMedia list, because None is not
// a member of any string list.
func TestExtractTracks_NullFormatMediumNeverIgnored(t *testing.T) {
	rel := Release{Media: []Medium{{
		Position: 1, Format: nil,
		Tracks: []Track{{ID: "t1", Position: 1, Number: "1", Title: "A", Recording: Recording{ID: "r1", Title: "A", ArtistCredit: []CreditPart{simpleCreditPart("a1", "Artist")}}}},
	}}}
	tun := matcher.DefaultTunables()
	tun.IgnoredMedia = []string{"CD", ""}
	tracks := extractTracks(&rel, tun)
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1 (a null-format medium is never ignored)", len(tracks))
	}
}

func loadSentinelTracks(t *testing.T, name string, tun matcher.Tunables) []builtTrack {
	t.Helper()
	rel := loadTestdata[Release](t, name)
	return extractTracks(&rel, tun)
}

// TestDeriveMedia_ZeroSurvivingTracksIsSentinel proves the first
// sentinel route: a release whose only track is a dropped video
// survives zero tracks, and deriveMedia returns the "Media" sentinel
// rather than "".
func TestDeriveMedia_ZeroSurvivingTracksIsSentinel(t *testing.T) {
	tun := matcher.DefaultTunables()
	tun.IgnoreVideoTracks = true
	tracks := loadSentinelTracks(t, "sentinel_zero_tracks.json", tun)
	if len(tracks) != 0 {
		t.Fatalf("got %d tracks, want 0", len(tracks))
	}
	if got := deriveMedia(tracks); got != "Media" {
		t.Errorf("deriveMedia(zero tracks) = %q, want %q", got, "Media")
	}
}

// TestDeriveMedia_MixedFormatsIsSentinel proves the second sentinel
// route: two distinct non-null formats across two media.
func TestDeriveMedia_MixedFormatsIsSentinel(t *testing.T) {
	tracks := loadSentinelTracks(t, "sentinel_mixed_formats.json", matcher.DefaultTunables())
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want 2", len(tracks))
	}
	if got := deriveMedia(tracks); got != "Media" {
		t.Errorf("deriveMedia(mixed formats) = %q, want %q", got, "Media")
	}
}

// TestDeriveMedia_SingleNullFormatIsNotSentinel proves the non-route
// revision 1 had backwards (fold, B5): a single medium whose format is
// null is the ONE distinct value in the set, so deriveMedia returns
// "", the falsy R1 zero-value mapping of beets' None, and declines the
// media penalty.
func TestDeriveMedia_SingleNullFormatIsNotSentinel(t *testing.T) {
	tracks := loadSentinelTracks(t, "sentinel_null_format.json", matcher.DefaultTunables())
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1", len(tracks))
	}
	if got := deriveMedia(tracks); got != "" {
		t.Errorf("deriveMedia(single null format) = %q, want %q", got, "")
	}
}

// TestDeriveMedia_NullAlongsideNonNullIsSentinel proves the route a
// reader least expects: None counts as one of the distinct values, so
// [null, "CD"] across two media has two distinct values and lands on
// the sentinel, not on "".
func TestDeriveMedia_NullAlongsideNonNullIsSentinel(t *testing.T) {
	tracks := loadSentinelTracks(t, "sentinel_null_and_cd.json", matcher.DefaultTunables())
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want 2", len(tracks))
	}
	if got := deriveMedia(tracks); got != "Media" {
		t.Errorf("deriveMedia([null, \"CD\"]) = %q, want %q", got, "Media")
	}
}
