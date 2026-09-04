package matcher

// AlbumInfo is the matcher's vocabulary for a candidate release: the
// fourteen fields the album-level distance path reads off beets'
// AlbumInfo (autotag/hooks.py:263-384), and nothing else. It carries no
// applied field, no path, and does no tag I/O: rung 3 computes no
// applied field and writes no file. 3b's translation layer populates an
// AlbumInfo from a real metadata-source response; this type only
// defines the shape and its zero value.
type AlbumInfo struct {
	// Artist is the release's artist credit (album_info.artist, set
	// through the Info base class).
	Artist string

	// Album is the release title (album_info.album, set through the
	// Info base class).
	Album string

	// VA reports whether this is a various-artists release
	// (album_info.va, hooks.py:381). When true, the album-level artist
	// component is skipped entirely (contract 8).
	VA bool

	// Media is the release's media type string, for example "CD"
	// (album_info.media, set through the Info base class).
	Media string

	// Mediums is the disc count (album_info.mediums, hooks.py:371).
	Mediums int

	// Year is the release year (album_info.year, hooks.py:382).
	Year int

	// OriginalYear is the original release year, distinct from Year
	// when this is a reissue (album_info.original_year,
	// hooks.py:375). Read by the year component only when
	// Tunables.Preferred.OriginalYear is set.
	OriginalYear int

	// Country is the release country code (album_info.country,
	// hooks.py:364).
	Country string

	// Label is the release label (album_info.label, hooks.py:369).
	Label string

	// CatalogNum is the label's catalog number (album_info.catalognum,
	// hooks.py:363).
	CatalogNum string

	// AlbumDisambig is the release disambiguation string
	// (album_info.albumdisambig, hooks.py:357).
	AlbumDisambig string

	// AlbumID is the release's source-specific identifier
	// (album_info.album_id, hooks.py:356).
	AlbumID string

	// DataSource names the metadata source this candidate came from,
	// for example "MusicBrainz" (album_info.data_source, set through
	// the Info base class).
	DataSource string

	// Tracks is the release's candidate track list, in the source's
	// own order (album_info.tracks, hooks.py:355).
	Tracks []TrackInfo
}

// TrackInfo is the matcher's vocabulary for one candidate track: the
// eight fields the track-level distance path reads off beets'
// TrackInfo (autotag/hooks.py:386-497), and nothing else. As with
// [AlbumInfo], this type carries no applied field and does no tag I/O.
type TrackInfo struct {
	// Title is the track title (track_info.title, hooks.py:491).
	Title string

	// Artist is the track's own artist credit, distinct from the
	// album's (track_info.artist, set through the Info base class).
	Artist string

	// Length is the track duration in seconds (track_info.length,
	// hooks.py:481).
	Length float64

	// Index is the track's position in the source's own numbering
	// (track_info.index, hooks.py:479).
	Index int

	// MediumIndex is the track's position within its disc
	// (track_info.medium_index, hooks.py:486).
	MediumIndex int

	// Medium is the disc number this track belongs to
	// (track_info.medium, hooks.py:485).
	Medium int

	// TrackID is the track's source-specific identifier
	// (track_info.track_id, hooks.py:493).
	TrackID string

	// DataSource names the metadata source this candidate came from,
	// for example "MusicBrainz" (track_info.data_source, set through
	// the Info base class).
	DataSource string
}

// ItemFields is the typed field bag the distance path reads off one
// user item: the eighteen fields get_most_common_tags
// (util/__init__.py:822-836) and track_distance
// (autotag/distance.py:394-424) touch between them, each in beets' own
// type. Artist and DataSource are read by both call sites and appear
// once here. A field outside this set is not read by the distance
// path; adding one is a finding, not a convenience.
//
// data_source is the one field whose true beets default is None rather
// than the empty string beets projects every other unset string field
// to (it lives in Item._types rather than Item._fields,
// library/models.py:66-70; contract 10's m6 note). The outcomes are
// identical through both consumers of it (add_data_source treats None
// as unset; plurality counts None as its own value), so this field is
// modeled with the same "" default as every other unset string field
// and is not the field [ProjectItemFields]'s own test uses to prove the
// typed-default rule.
type ItemFields struct {
	Artist        string
	Album         string
	AlbumArtist   string
	Year          int
	DiscTotal     int
	MBAlbumID     string
	Label         string
	Barcode       string
	CatalogNum    string
	Country       string
	Media         string
	AlbumDisambig string
	DataSource    string
	Length        float64
	Title         string
	Track         int
	MBTrackID     string
	Disc          int
}

// The beets item field names [ProjectItemFields] reads each
// [ItemFields] member from.
const (
	fieldArtist        = "artist"
	fieldAlbum         = "album"
	fieldAlbumArtist   = "albumartist"
	fieldYear          = "year"
	fieldDiscTotal     = "disctotal"
	fieldMBAlbumID     = "mb_albumid"
	fieldLabel         = "label"
	fieldBarcode       = "barcode"
	fieldCatalogNum    = "catalognum"
	fieldCountry       = "country"
	fieldMedia         = "media"
	fieldAlbumDisambig = "albumdisambig"
	fieldDataSource    = "data_source"
	fieldLength        = "length"
	fieldTitle         = "title"
	fieldTrack         = "track"
	fieldMBTrackID     = "mb_trackid"
	fieldDisc          = "disc"
)

// ProjectItemFields builds an [ItemFields] from raw, an untyped field
// bag shaped exactly like a corpus case's decoded Input.Items entry
// (json.Unmarshal into map[string]any, so an integer value may arrive
// as float64). It reproduces beets' typed-default projection (contract
// 10): a beets Item.get returns the field's own type's zero value for
// a field the underlying record never set, not an absent marker, and a
// field this function cannot read as its declared type is treated the
// same as an absent one. This function reads no corpus case itself
// (the cross-check against a real committed case is Task 6's); its own
// test fixture is hand-authored inline in info_test.go and reads no
// corpus case.
func ProjectItemFields(raw map[string]any) ItemFields {
	return ItemFields{
		Artist:        stringField(raw, fieldArtist),
		Album:         stringField(raw, fieldAlbum),
		AlbumArtist:   stringField(raw, fieldAlbumArtist),
		Year:          intField(raw, fieldYear),
		DiscTotal:     intField(raw, fieldDiscTotal),
		MBAlbumID:     stringField(raw, fieldMBAlbumID),
		Label:         stringField(raw, fieldLabel),
		Barcode:       stringField(raw, fieldBarcode),
		CatalogNum:    stringField(raw, fieldCatalogNum),
		Country:       stringField(raw, fieldCountry),
		Media:         stringField(raw, fieldMedia),
		AlbumDisambig: stringField(raw, fieldAlbumDisambig),
		DataSource:    stringField(raw, fieldDataSource),
		Length:        floatField(raw, fieldLength),
		Title:         stringField(raw, fieldTitle),
		Track:         intField(raw, fieldTrack),
		MBTrackID:     stringField(raw, fieldMBTrackID),
		Disc:          intField(raw, fieldDisc),
	}
}

// stringField returns raw[key] as a string, or "" when key is absent
// or holds a non-string value.
func stringField(raw map[string]any, key string) string {
	s, _ := raw[key].(string)
	return s
}

// intField returns raw[key] as an int, or 0 when key is absent or
// holds neither an int nor a float64 (encoding/json decodes a JSON
// number into map[string]any as float64).
func intField(raw map[string]any, key string) int {
	switch v := raw[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

// floatField returns raw[key] as a float64, or 0 when key is absent or
// holds neither a float64 nor an int.
func floatField(raw map[string]any, key string) float64 {
	switch v := raw[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0
	}
}
