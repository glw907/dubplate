package matcher

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"
)

// TestDistanceForAlbumComposition hand-authors three whole-distance
// scenarios on a small synthetic album. These are not corpus cases and
// read nothing under corpus/; every want below was measured by calling
// beets.autotag.distance.distance in the pinned venv (tools/venv/.venv,
// beets 2.13.1) on exactly these inputs, config defaults throughout
// (no preferred media/countries, no original-year preference), and the
// comparison is exact rather than tolerant.
func TestDistanceForAlbumComposition(t *testing.T) {
	evalDate := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	t.Run("clean pairing: every component declines or scores zero", func(t *testing.T) {
		trackA := TrackInfo{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Index: 1, MediumIndex: 1, Medium: 1, TrackID: "mbid-track-1",
			DataSource: "MusicBrainz",
		}
		trackB := TrackInfo{
			Title: "Quiet Season", Artist: "Northern Lines", Length: 180,
			Index: 2, MediumIndex: 2, Medium: 1, TrackID: "mbid-track-2",
			DataSource: "MusicBrainz",
		}
		info := &AlbumInfo{
			Artist: "Northern Lines", Album: "Ember Wharf", VA: false,
			Media: "CD", Mediums: 1, Year: 2020, Country: "US",
			Label: "Quiet Season Records", CatalogNum: "QS001",
			AlbumID: "mbid-album-1", DataSource: "MusicBrainz",
			Tracks: []TrackInfo{trackA, trackB},
		}
		itemA := ItemFields{
			Artist: "Northern Lines", Album: "Ember Wharf", AlbumArtist: "Northern Lines",
			Year: 2020, DiscTotal: 1, MBAlbumID: "mbid-album-1", Label: "Quiet Season Records",
			CatalogNum: "QS001", Country: "US", Media: "CD", DataSource: "MusicBrainz",
			Length: 200, Title: "Ember Wharf", Track: 1, MBTrackID: "mbid-track-1", Disc: 1,
		}
		itemB := ItemFields{
			Artist: "Northern Lines", Album: "Ember Wharf", AlbumArtist: "Northern Lines",
			Year: 2020, DiscTotal: 1, MBAlbumID: "mbid-album-1", Label: "Quiet Season Records",
			CatalogNum: "QS001", Country: "US", Media: "CD", DataSource: "MusicBrainz",
			Length: 180, Title: "Quiet Season", Track: 2, MBTrackID: "mbid-track-2", Disc: 1,
		}
		items := []ItemFields{itemA, itemB}
		pairs := []AlbumPair{{Item: itemA, Info: &trackA}, {Item: itemB, Info: &trackB}}

		dist := DistanceForAlbum(items, info, pairs, DefaultTunables(), evalDate)

		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "artist", "album", "media", "mediums", "year",
			"country", "label", "catalognum", "album_id", "tracks")
		if len(dist.Tracks) != 2 {
			t.Fatalf("len(dist.Tracks) = %d, want 2", len(dist.Tracks))
		}
	})

	t.Run("missing track and unmatched item: counting components inflate the denominator", func(t *testing.T) {
		trackA := TrackInfo{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Index: 1, MediumIndex: 1, Medium: 1, TrackID: "mbid-track-1",
			DataSource: "MusicBrainz",
		}
		trackB := TrackInfo{
			Title: "Second Light", Artist: "Northern Lines", Length: 210,
			Index: 2, MediumIndex: 2, Medium: 1, TrackID: "mbid-track-3",
			DataSource: "MusicBrainz",
		}
		info := &AlbumInfo{
			Artist: "Northern Lines", Album: "Ember Wharf", VA: false,
			Media: "CD", Mediums: 1, Year: 2020, Country: "US",
			Label: "Quiet Season Records", CatalogNum: "QS001",
			AlbumID: "mbid-album-1", DataSource: "MusicBrainz",
			// trackB has no matching item: proves missing_tracks.
			Tracks: []TrackInfo{trackA, trackB},
		}
		itemA := ItemFields{
			Artist: "Northern Lines", Album: "Ember Wharf", AlbumArtist: "Northern Lines",
			Year: 2020, DiscTotal: 1, MBAlbumID: "mbid-album-1", Label: "Quiet Season Records",
			CatalogNum: "QS001", Country: "US", Media: "CD", DataSource: "MusicBrainz",
			Length: 200, Title: "Ember Wharf", Track: 1, MBTrackID: "mbid-track-1", Disc: 1,
		}
		// itemC has no matching candidate track: proves unmatched_tracks.
		itemC := ItemFields{
			Artist: "Northern Lines", Album: "Ember Wharf", AlbumArtist: "Northern Lines",
			Year: 2020, DiscTotal: 1, MBAlbumID: "mbid-album-1", Label: "Quiet Season Records",
			CatalogNum: "QS001", Country: "US", Media: "CD", DataSource: "MusicBrainz",
			Length: 999, Title: "Unrelated Bonus Track", Track: 9, MBTrackID: "mbid-track-unmatched", Disc: 1,
		}
		items := []ItemFields{itemA, itemC}
		pairs := []AlbumPair{{Item: itemA, Info: &trackA}}

		dist := DistanceForAlbum(items, info, pairs, DefaultTunables(), evalDate)

		wantDistance(t, dist, 0.07894736842105263)
		wantAddedKeys(t, dist, "artist", "album", "media", "mediums", "year",
			"country", "label", "catalognum", "album_id", "tracks",
			"missing_tracks", "unmatched_tracks")
		perKey := dist.PerKey()
		if got := perKey["missing_tracks"]; got != 0.04736842105263158 {
			t.Errorf("PerKey()[missing_tracks] = %v, want 0.04736842105263158", got)
		}
		if got := perKey["unmatched_tracks"]; got != 0.031578947368421054 {
			t.Errorf("PerKey()[unmatched_tracks] = %v, want 0.031578947368421054", got)
		}
		if len(dist.Tracks) != 1 {
			t.Fatalf("len(dist.Tracks) = %d, want 1", len(dist.Tracks))
		}
	})

	t.Run("VA album: the artist skip and the inclArtist=true path through tracks", func(t *testing.T) {
		trackA := TrackInfo{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Index: 1, MediumIndex: 1, Medium: 1, TrackID: "mbid-track-1",
			DataSource: "MusicBrainz",
		}
		// trackB's artist diverges from itemB's: on a non-VA album this
		// would never surface here (the album-level artist component
		// would already have failed), but a VA release skips that
		// component and instead passes inclArtist=true into the
		// per-track distance, so track_artist fires inside dist.Tracks.
		trackB := TrackInfo{
			Title: "Quiet Season", Artist: "Cold Harbor", Length: 180,
			Index: 2, MediumIndex: 2, Medium: 1, TrackID: "mbid-track-2",
			DataSource: "MusicBrainz",
		}
		info := &AlbumInfo{
			Artist: "Various Artists", Album: "Ember Wharf", VA: true,
			Media: "CD", Mediums: 1, Year: 2020, Country: "US",
			Label: "Quiet Season Records", CatalogNum: "QS001",
			AlbumID: "mbid-album-1", DataSource: "MusicBrainz",
			Tracks: []TrackInfo{trackA, trackB},
		}
		itemA := ItemFields{
			Artist: "Northern Lines", Album: "Ember Wharf", AlbumArtist: "Various Artists",
			Year: 2020, DiscTotal: 1, MBAlbumID: "mbid-album-1", Label: "Quiet Season Records",
			CatalogNum: "QS001", Country: "US", Media: "CD", DataSource: "MusicBrainz",
			Length: 200, Title: "Ember Wharf", Track: 1, MBTrackID: "mbid-track-1", Disc: 1,
		}
		itemB := ItemFields{
			Artist: "Northern Lines", Album: "Ember Wharf", AlbumArtist: "Various Artists",
			Year: 2020, DiscTotal: 1, MBAlbumID: "mbid-album-1", Label: "Quiet Season Records",
			CatalogNum: "QS001", Country: "US", Media: "CD", DataSource: "MusicBrainz",
			Length: 180, Title: "Quiet Season", Track: 2, MBTrackID: "mbid-track-2", Disc: 1,
		}
		items := []ItemFields{itemA, itemB}
		pairs := []AlbumPair{{Item: itemA, Info: &trackA}, {Item: itemB, Info: &trackB}}

		dist := DistanceForAlbum(items, info, pairs, DefaultTunables(), evalDate)

		wantDistance(t, dist, 0.013320013320013322)
		// "artist" is absent: the VA skip, not a zero-scoring add.
		wantAddedKeys(t, dist, "album", "media", "mediums", "year", "country",
			"label", "catalognum", "album_id", "tracks")
		if len(dist.Tracks) != 2 {
			t.Fatalf("len(dist.Tracks) = %d, want 2", len(dist.Tracks))
		}
		if got := dist.Tracks[0].Dist.Distance(); got != 0 {
			t.Errorf("dist.Tracks[0].Dist.Distance() = %v, want 0 (identical pair)", got)
		}
		if got := dist.Tracks[1].Dist.Distance(); got != 0.1098901098901099 {
			t.Errorf("dist.Tracks[1].Dist.Distance() = %v, want 0.1098901098901099 (track_artist fires)", got)
		}
		if !slices.Contains(dist.Tracks[1].Dist.AddedKeys(), "track_artist") {
			t.Errorf("dist.Tracks[1].Dist.AddedKeys() = %v, want it to contain track_artist", dist.Tracks[1].Dist.AddedKeys())
		}
	})
}

func TestAddAlbumArtist(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		va     bool
		likely string
		info   string
		want   []float64
	}{
		{"declines entirely on a various-artists release", true, "Northern Lines", "Various Artists", nil},
		{"adds zero for identical artists", false, "Northern Lines", "Northern Lines", []float64{0}},
		{"adds a nonzero string distance for different artists", false, "Northern Lines", "Ember Wharf", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Artist: tt.likely}
			info := &AlbumInfo{VA: tt.va, Artist: tt.info}
			addAlbumArtist(dist, likelies, info)

			got := dist.penalties["artist"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumArtist(va=%v, %q, %q) added %v, want %v", tt.va, tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

// album carries no fire condition, unlike most album components, so
// this test's decline proof is the edge case a fire-conditioned
// component would use to decline: both titles empty. addAlbumAlbum
// still adds here, which is the point.
func TestAddAlbumAlbum(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"always adds, even when both titles are empty", "", "", []float64{0}},
		{"adds zero for identical titles", "Harbour Sound", "Harbour Sound", []float64{0}},
		{"adds a nonzero string distance for different titles", "Ember Wharf", "Northern Lines", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Album: tt.likely}
			info := &AlbumInfo{Album: tt.info}
			addAlbumAlbum(dist, likelies, info)

			got := dist.penalties["album"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumAlbum(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

// TestAddAlbumMedia proves the media pattern anchor: beets matches a
// compiled option with re.Pattern.match, which anchors at position 0
// (distance.py:265-267), so a pattern later in the value never matches.
func TestAddAlbumMedia(t *testing.T) {
	tests := []struct {
		name      string
		preferred []string
		likely    string
		info      string
		want      []float64
	}{
		{"declines when the candidate carries no media", nil, "CD", "", nil},
		{"declines when neither preferred patterns nor a likely media are set", nil, "", "CD", nil},
		{"equality: current media matches", nil, "CD", "CD", []float64{0}},
		{"equality: current media mismatches", nil, "CD", "Vinyl", []float64{1}},
		{"priority: the first pattern matches", []string{"CD", "Vinyl"}, "", "CD", []float64{0}},
		{"priority: the second pattern matches", []string{"CD", "Vinyl"}, "", "Vinyl", []float64{0.5}},
		{"priority: a disc-count prefix is tolerated", []string{"CD", "Vinyl"}, "", "2xCD", []float64{0}},
		{"priority: case-insensitive", []string{"cd"}, "", "CD", []float64{0}},
		{"priority: nothing matches", []string{"CD", "Vinyl"}, "", "Cassette", []float64{1}},
		{"priority: a pattern later in the value does not match", []string{"CD"}, "", "Deluxe CD", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(DefaultTunables())
			likelies := Likelies{Media: tt.likely}
			info := &AlbumInfo{Media: tt.info}
			addAlbumMedia(dist, tt.preferred, likelies, info)

			got := dist.penalties["media"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumMedia(preferred=%v, likely=%q, info=%q) added %v, want %v", tt.preferred, tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

func TestAddAlbumMediums(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name      string
		discTotal int
		mediums   int
		want      []float64
	}{
		{"declines when the item set carries no disc total", 0, 2, nil},
		{"declines when the candidate carries no medium count", 2, 0, nil},
		{"matching disc counts add zero", 2, 2, []float64{0}},
		{"a one-disc difference adds one unit", 1, 2, []float64{1}},
		{"a two-disc difference adds two units", 1, 3, []float64{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{DiscTotal: tt.discTotal}
			info := &AlbumInfo{Mediums: tt.mediums}
			addAlbumMediums(dist, likelies, info)

			got := dist.penalties["mediums"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumMediums(discTotal=%d, mediums=%d) added %v, want %v", tt.discTotal, tt.mediums, got, tt.want)
			}
		})
	}
}

// TestAddAlbumYearPreferredOriginalYear grades branch 1, the
// preferred.original_year path, off a hand-authored table
// (testdata/album_year_preferred_original.json) because pinned beets'
// corpus carries no case that sets preferred.original_year. Every
// "want" is hand-computed from the formula AddRatio(|year - original|,
// |evalYear - original|), original = originalYear or 1889.
func TestAddAlbumYearPreferredOriginalYear(t *testing.T) {
	type fixture struct {
		Name         string  `json:"name"`
		Year         int     `json:"year"`
		OriginalYear int     `json:"originalYear"`
		EvalYear     int     `json:"evalYear"`
		Want         float64 `json:"want"`
	}

	raw, err := os.ReadFile("testdata/album_year_preferred_original.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("testdata/album_year_preferred_original.json carries no fixtures")
	}

	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			dist := NewDistance(DefaultTunables())
			info := &AlbumInfo{Year: f.Year, OriginalYear: f.OriginalYear}
			addAlbumYear(dist, true, Likelies{}, info, f.EvalYear)

			got := dist.penalties["year"]
			if want := []float64{f.Want}; !slices.Equal(got, want) {
				t.Errorf("addAlbumYear(year=%d, originalYear=%d, evalYear=%d) added %v, want %v",
					f.Year, f.OriginalYear, f.EvalYear, got, want)
			}
		})
	}
}

func TestAddAlbumYear(t *testing.T) {
	tun := DefaultTunables()
	const evalYear = 2025

	tests := []struct {
		name         string
		likelyYear   int
		infoYear     int
		originalYear int
		want         []float64
	}{
		{"declines when the item set carries no likely year", 0, 1975, 0, nil},
		{"declines when the candidate carries no release year", 1980, 0, 0, nil},
		{"matching release years add zero", 1975, 1975, 0, []float64{0}},
		{"matching original years add zero even with a different release year", 1970, 1975, 1970, []float64{0}},
		{"a candidate with an original year ratios plainly", 1980, 1975, 1970, []float64{5.0 / 55.0}},
		{"no original year is a flat penalty", 1980, 1975, 0, []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Year: tt.likelyYear}
			info := &AlbumInfo{Year: tt.infoYear, OriginalYear: tt.originalYear}
			addAlbumYear(dist, false, likelies, info, evalYear)

			got := dist.penalties["year"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumYear(likely=%d, year=%d, original=%d) added %v, want %v",
					tt.likelyYear, tt.infoYear, tt.originalYear, got, tt.want)
			}
		})
	}
}

// TestAddAlbumYearPreferredOriginalYearTakesPrecedence proves branch
// 1 is tried before branch 2: with preferOriginal set, even a
// likely-year match that branch 2 would score zero for is ignored.
func TestAddAlbumYearPreferredOriginalYearTakesPrecedence(t *testing.T) {
	dist := NewDistance(DefaultTunables())
	likelies := Likelies{Year: 1975}
	info := &AlbumInfo{Year: 1975, OriginalYear: 1970}
	addAlbumYear(dist, true, likelies, info, 2025)

	got := dist.penalties["year"]
	want := []float64{5.0 / 55.0}
	if !slices.Equal(got, want) {
		t.Errorf("addAlbumYear with preferOriginal added %v, want %v (branch 1, not branch 2's zero)", got, want)
	}
}

// TestAddAlbumCountry proves the country pattern anchor the way
// TestAddAlbumMedia proves the media one.
func TestAddAlbumCountry(t *testing.T) {
	tests := []struct {
		name      string
		preferred []string
		likely    string
		info      string
		want      []float64
	}{
		{"declines when neither preferred patterns nor a likely country produce a match", nil, "", "GB", nil},
		{"declines when the candidate carries no country and nothing else fires", nil, "US", "", nil},
		{"string: matching codes add zero", nil, "GB", "GB", []float64{0}},
		{"string: mismatched codes add a nonzero distance", nil, "US", "GB", []float64{1}},
		{"priority: the first pattern matches", []string{"GB", "US"}, "", "GB", []float64{0}},
		{"priority: the second pattern matches", []string{"GB", "US"}, "", "US", []float64{0.5}},
		{"priority: case-insensitive", []string{"gb"}, "", "GB", []float64{0}},
		{"priority: nothing matches", []string{"GB", "US"}, "", "FR", []float64{1}},
		{"priority takes precedence over the string fallback", []string{"GB"}, "US", "GB", []float64{0}},
		{"priority: a pattern later in the value does not match", []string{"GB"}, "", "XGB", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(DefaultTunables())
			likelies := Likelies{Country: tt.likely}
			info := &AlbumInfo{Country: tt.info}
			addAlbumCountry(dist, tt.preferred, likelies, info)

			got := dist.penalties["country"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumCountry(preferred=%v, likely=%q, info=%q) added %v, want %v", tt.preferred, tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

func TestAddAlbumLabel(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely label", "", "Cedarwood Records", nil},
		{"declines when the candidate carries no label", "Cedarwood Records", "", nil},
		{"adds zero for identical labels", "Cedarwood Records", "Cedarwood Records", []float64{0}},
		{"adds a nonzero string distance for different labels", "Cedarwood Records", "Cedarwood Recordings", []float64{0.15789473684210525}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Label: tt.likely}
			info := &AlbumInfo{Label: tt.info}
			addAlbumLabel(dist, likelies, info)

			got := dist.penalties["label"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumLabel(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

func TestAddAlbumCatalogNum(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely catalog number", "", "CAT-001", nil},
		{"declines when the candidate carries no catalog number", "CAT-001", "", nil},
		{"adds zero for identical catalog numbers", "CAT-001", "CAT-001", []float64{0}},
		{"adds a nonzero string distance for different catalog numbers", "CAT-001", "CAT-002", []float64{0.16666666666666666}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{CatalogNum: tt.likely}
			info := &AlbumInfo{CatalogNum: tt.info}
			addAlbumCatalogNum(dist, likelies, info)

			got := dist.penalties["catalognum"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumCatalogNum(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

func TestAddAlbumDisambig(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely disambiguation", "", "Deluxe Edition", nil},
		{"declines when the candidate carries no disambiguation", "Deluxe Edition", "", nil},
		{"adds zero for identical disambiguation strings", "Deluxe Edition", "Deluxe Edition", []float64{0}},
		{"adds a nonzero string distance for different disambiguation strings", "Deluxe Edition", "Deluxe Reissue", []float64{0.46153846153846156}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{AlbumDisambig: tt.likely}
			info := &AlbumInfo{AlbumDisambig: tt.info}
			addAlbumDisambig(dist, likelies, info)

			got := dist.penalties["albumdisambig"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumDisambig(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

func TestAddAlbumID(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely album ID", "", "mb-1", nil},
		{"matching IDs add zero", "mb-1", "mb-1", []float64{0}},
		{"mismatched IDs add one", "mb-1", "mb-2", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{MBAlbumID: tt.likely}
			info := &AlbumInfo{AlbumID: tt.info}
			addAlbumID(dist, likelies, info)

			got := dist.penalties["album_id"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumID(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}

func TestAddAlbumTracks(t *testing.T) {
	tun := DefaultTunables()

	t.Run("declines (adds nothing) with zero pairs", func(t *testing.T) {
		dist := NewDistance(tun)
		addAlbumTracks(dist, tun, false, nil)

		if got := dist.penalties["tracks"]; len(got) != 0 {
			t.Errorf("addAlbumTracks(no pairs) added %v, want nothing", got)
		}
		if len(dist.Tracks) != 0 {
			t.Errorf("addAlbumTracks(no pairs) recorded %d sub-distances, want 0", len(dist.Tracks))
		}
	})

	t.Run("adds one normalized track distance per pair and records the sub-distance", func(t *testing.T) {
		dist := NewDistance(tun)
		pairs := []AlbumPair{
			{
				Item: ItemFields{Title: "Ember Wharf"},
				Info: &TrackInfo{Title: "Ember Wharf"},
			},
			{
				Item: ItemFields{Title: "A Quiet Season"},
				Info: &TrackInfo{Title: "Ember Wharf"},
			},
		}
		addAlbumTracks(dist, tun, false, pairs)

		got := dist.penalties["tracks"]
		if len(got) != 2 {
			t.Fatalf("addAlbumTracks(2 pairs) added %v, want 2 entries", got)
		}
		if got[0] != 0 {
			t.Errorf("first pair (identical) added %v, want 0", got[0])
		}
		if got[1] == 0 {
			t.Errorf("second pair (mismatched title) added 0, want a nonzero distance")
		}
		if len(dist.Tracks) != 2 {
			t.Fatalf("addAlbumTracks(2 pairs) recorded %d sub-distances, want 2", len(dist.Tracks))
		}
		if dist.Tracks[0].Info != pairs[0].Info || dist.Tracks[1].Info != pairs[1].Info {
			t.Error("addAlbumTracks recorded sub-distances against the wrong TrackInfo")
		}
	})

	// The album-level tracks component threads album_info.VA straight
	// through as track_distance's inclArtist, unnegated
	// (distance.py:529: `track_distance(item, track, album_info.va)`).
	// On a non-VA release the album-level artist component already
	// grades the artist once, so the per-track fold skips it; on a VA
	// release there is no single album artist to grade, so the
	// per-track fold is where a mismatched track artist gets caught.
	t.Run("inclArtist follows album_info.VA unnegated", func(t *testing.T) {
		pairs := []AlbumPair{
			{
				Item: ItemFields{Title: "Ember Wharf", Artist: "Quiet Season"},
				Info: &TrackInfo{Title: "Ember Wharf", Artist: "Northern Lines"},
			},
		}

		nonVA := NewDistance(tun)
		addAlbumTracks(nonVA, tun, false, pairs)
		if got := nonVA.penalties["tracks"]; len(got) != 1 || got[0] != 0 {
			t.Errorf("non-VA release added %v, want [0] (artist not re-graded per track)", got)
		}

		va := NewDistance(tun)
		addAlbumTracks(va, tun, true, pairs)
		if got := va.penalties["tracks"]; len(got) != 1 || got[0] == 0 {
			t.Errorf("VA release added %v, want a nonzero distance (artist graded per track)", got)
		}
	})
}

func TestAddAlbumMissingTracks(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		infoTracks int
		pairs      int
		want       []float64
	}{
		{"declines when every candidate track is paired", 2, 2, nil},
		{"declines when the pairing covers every candidate track and more", 1, 2, nil},
		{"one 1.0 penalty per unpaired candidate track", 3, 1, []float64{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			info := &AlbumInfo{Tracks: make([]TrackInfo, tt.infoTracks)}
			pairs := make([]AlbumPair, tt.pairs)
			addAlbumMissingTracks(dist, info, pairs)

			got := dist.penalties["missing_tracks"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumMissingTracks(tracks=%d, pairs=%d) added %v, want %v", tt.infoTracks, tt.pairs, got, tt.want)
			}
		})
	}
}

func TestAddAlbumUnmatchedTracks(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name  string
		items int
		pairs int
		want  []float64
	}{
		{"declines when every item is paired", 2, 2, nil},
		{"declines when the pairing covers every item and more", 1, 2, nil},
		{"one 1.0 penalty per unmatched item", 3, 1, []float64{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			items := make([]ItemFields, tt.items)
			pairs := make([]AlbumPair, tt.pairs)
			addAlbumUnmatchedTracks(dist, items, pairs)

			got := dist.penalties["unmatched_tracks"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumUnmatchedTracks(items=%d, pairs=%d) added %v, want %v", tt.items, tt.pairs, got, tt.want)
			}
		})
	}
}

// TestAddAlbumDataSource walks AddDataSource's three-way condition in
// all four states, mirroring TestAddTrackDataSource: this component
// carries no fire condition of its own and wires likelies.DataSource
// and info.DataSource straight into AddDataSource.
func TestAddAlbumDataSource(t *testing.T) {
	twoSources := map[string]float64{"MusicBrainz": 0.5, "Spotify": 0.75}
	oneSource := map[string]float64{"MusicBrainz": 0.5}

	tests := []struct {
		name    string
		sources map[string]float64
		before  string
		after   string
		want    []float64
	}{
		{"sources match", twoSources, "MusicBrainz", "MusicBrainz", nil},
		{"a mismatch against a set before", twoSources, "MusicBrainz", "Spotify", []float64{0.75}},
		{"no before and a single loaded source", oneSource, "", "Spotify", nil},
		{"no before and several loaded sources", twoSources, "", "Spotify", []float64{0.75}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.DataSourcePenalties = tt.sources
			dist := NewDistance(tun)
			likelies := Likelies{DataSource: tt.before}
			info := &AlbumInfo{DataSource: tt.after}
			addAlbumDataSource(dist, likelies, info)

			got := dist.penalties["data_source"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumDataSource(%q, %q) added %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
