package matcher

import (
	"slices"
	"testing"
	"time"
)

// TestAlbumComponentsComposeInFixedOrder pins contract 8's album order
// directly, mirroring track.go's TestTrackDistanceComposesInFixedOrder.
//
// The mutation that proves this fires: transposing two adjacent
// entries in albumComponents (media and mediums) makes this test fail
// with
//
//	albumComponents order = [artist album mediums media year country label catalognum albumdisambig album_id tracks missing_tracks unmatched_tracks data_source], want [artist album media mediums year country label catalognum albumdisambig album_id tracks missing_tracks unmatched_tracks data_source]
//
// confirmed by hand during development and reverted before commit; this
// test is what would have caught it.
func TestAlbumComponentsComposeInFixedOrder(t *testing.T) {
	want := []string{
		"artist", "album", "media", "mediums", "year", "country", "label",
		"catalognum", "albumdisambig", "album_id", "tracks", "missing_tracks",
		"unmatched_tracks", "data_source",
	}
	got := make([]string, len(albumComponents))
	for i, c := range albumComponents {
		got[i] = c.Name
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("albumComponents order = %v, want %v", got, want)
		}
	}
}

// TestComponentKeysAreWeighted is Key's reader (score.go's doc
// comment): every track and album component's Key must name an entry
// in Tunables.DistanceWeights, the twenty-key set the album and track
// key sets share "data_source" across (m2's twenty-key resolution). A
// component whose Key does not match a configured weight would score
// silently against an unweighted key, and this test is what would
// catch it: mutating trackDataSourceComponent's Key to "datasource"
// (the underscore dropped) fails here with `component "data_source"
// (track) has key "datasource", not in Tunables.DistanceWeights`.
func TestComponentKeysAreWeighted(t *testing.T) {
	weights := DefaultTunables().DistanceWeights
	for _, c := range trackComponents {
		if _, ok := weights[c.Key]; !ok {
			t.Errorf("component %q (track) has key %q, not in Tunables.DistanceWeights", c.Name, c.Key)
		}
	}
	for _, c := range albumComponents {
		if _, ok := weights[c.Key]; !ok {
			t.Errorf("component %q (album) has key %q, not in Tunables.DistanceWeights", c.Name, c.Key)
		}
	}
}

// TestDistanceForAlbumComposition hand-authors three whole-distance
// scenarios on a small synthetic album. These are not corpus cases and
// read nothing under corpus/ (whole-case agreement against the corpus
// is tier 1's, at 3d); every want below was measured by calling
// beets.autotag.distance.distance in the pinned venv (tools/venv/.venv,
// beets 2.13.1) on exactly these inputs, config defaults throughout
// (no preferred media/countries, no original-year preference), and the
// comparison is exact rather than tolerant (contract 12).
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
