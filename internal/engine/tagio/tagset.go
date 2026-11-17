package tagio

// Field names one of TagSet's 69 writable members, spelled the way
// corpus/README.md's handoff section and INV:315-385 both spell it
// (mediafile's own field name, snake_case). It is the identifier
// the field table's rows key on and the value WriteOrder yields.
type Field string

// The 69 writable fields, one constant per TagSet member. Names come
// from corpus/README.md:2381 onward, the handoff section C1 names as
// the single source of record; INV:315-385 supplies each field's
// Vorbis keys and mediafile citation, never its name.
const (
	FieldAcoustIDFingerprint Field = "acoustid_fingerprint"
	FieldAcoustIDID          Field = "acoustid_id"
	FieldAlbum               Field = "album"
	FieldAlbumartist         Field = "albumartist"
	FieldAlbumartistCredit   Field = "albumartist_credit"
	FieldAlbumartistSort     Field = "albumartist_sort"
	FieldAlbumartists        Field = "albumartists"
	FieldAlbumartistsCredit  Field = "albumartists_credit"
	FieldAlbumartistsSort    Field = "albumartists_sort"
	FieldAlbumdisambig       Field = "albumdisambig"
	FieldAlbumstatus         Field = "albumstatus"
	FieldAlbumtype           Field = "albumtype"
	FieldAlbumtypes          Field = "albumtypes"
	FieldArrangers           Field = "arrangers"
	FieldArtist              Field = "artist"
	FieldArtistCredit        Field = "artist_credit"
	FieldArtistSort          Field = "artist_sort"
	FieldArtists             Field = "artists"
	FieldArtistsCredit       Field = "artists_credit"
	FieldArtistsSort         Field = "artists_sort"
	FieldAsin                Field = "asin"
	FieldBarcode             Field = "barcode"
	FieldBpm                 Field = "bpm"
	FieldCatalognum          Field = "catalognum"
	FieldComments            Field = "comments"
	FieldComp                Field = "comp"
	FieldComposerSort        Field = "composer_sort"
	FieldComposers           Field = "composers"
	FieldCountry             Field = "country"
	FieldYear                Field = "year"
	FieldMonth               Field = "month"
	FieldDay                 Field = "day"
	FieldDisc                Field = "disc"
	FieldDisctitle           Field = "disctitle"
	FieldDisctotal           Field = "disctotal"
	FieldEncoder             Field = "encoder"
	FieldGenres              Field = "genres"
	FieldGrouping            Field = "grouping"
	FieldInitialKey          Field = "initial_key"
	FieldIsrc                Field = "isrc"
	FieldLabel               Field = "label"
	FieldLanguage            Field = "language"
	FieldLyricists           Field = "lyricists"
	FieldLyrics              Field = "lyrics"
	FieldMBAlbumartistid     Field = "mb_albumartistid"
	FieldMBAlbumartistids    Field = "mb_albumartistids"
	FieldMBAlbumid           Field = "mb_albumid"
	FieldMBArtistid          Field = "mb_artistid"
	FieldMBArtistids         Field = "mb_artistids"
	FieldMBReleasegroupid    Field = "mb_releasegroupid"
	FieldMBReleasetrackid    Field = "mb_releasetrackid"
	FieldMBTrackid           Field = "mb_trackid"
	FieldMBWorkid            Field = "mb_workid"
	FieldMedia               Field = "media"
	FieldOriginalYear        Field = "original_year"
	FieldOriginalMonth       Field = "original_month"
	FieldOriginalDay         Field = "original_day"
	FieldR128AlbumGain       Field = "r128_album_gain"
	FieldR128TrackGain       Field = "r128_track_gain"
	FieldRemixers            Field = "remixers"
	FieldRgAlbumGain         Field = "rg_album_gain"
	FieldRgAlbumPeak         Field = "rg_album_peak"
	FieldRgTrackGain         Field = "rg_track_gain"
	FieldRgTrackPeak         Field = "rg_track_peak"
	FieldScript              Field = "script"
	FieldSubtitle            Field = "subtitle"
	FieldTitle               Field = "title"
	FieldTrack               Field = "track"
	FieldTracktotal          Field = "tracktotal"
)

// TagSet is tagio's own vocabulary for the 69 fields mediafile 0.17.0
// (the version corpus/manifest.json pins) writes to a FLAC container:
// the intersection of beets' Item._media_fields and
// MediaFile.sorted_fields(), matching corpus/README.md's handoff
// section key for key (C1). No consumer's type appears here (ruling
// 3); a reader or writer that wants a beets Item or a matcher type
// converts at its own boundary.
//
// Every member carries a presence bit rather than a Go zero value,
// because a Go string's zero value and a Go slice's zero value are
// the same in Go and are not the same on disk: MediaFile.update
// deletes a key on None and sets it on everything else, so a string
// field's empty value writes a comment with an empty value while a
// list field's empty value writes no comment at all (C2). A string
// member is *string: nil is absent, a non-nil pointer to "" is
// present-but-empty, and a non-nil pointer to a non-empty string is a
// value. A list member is a plain slice, not a pointer: a nil slice
// is absent and a non-nil zero-length slice is present-but-empty,
// which is idiomatic Go's own nil/non-nil-empty distinction rather
// than a pointer indirection, so EmptyList's non-nil []string{} carries
// meaning a bare append never would. A numeric or bool member is a
// pointer (*int, *float64, *bool) with only two reachable states, nil
// absent and a value, because mediafile's ledger records no "empty
// int" or "empty bool": zero is a value like any other.
//
// Both nil-vs-empty distinctions are read and written directly, with
// no second lookup: a caller checking presence reads the one field
// TagSet already exposes for it.
//
// Every member carries a JSON tag naming the ledger's own field name
// (its Field constant's value), with no omitempty (R11): the
// conformance instrument fills a TagSet by unmarshalling a ledger
// row's filtered applied-fields object straight into it and projects
// a read-back TagSet by marshalling it, and every ledger row carries
// all 69 keys, absent values included as explicit nulls, so an
// omitted member would drop an absent field from the round trip.
// TestTagSetJSONTagsTieToFieldConstants ties every tag to the Field
// constant that names it (TestTagSetJSONTagsCarryNoOmitempty grades
// the no-omitempty half), so a wrong tag fails loudly rather than
// silently swapping two fields' values while both sides of a
// comparison agree.
type TagSet struct {
	AcoustIDFingerprint *string  `json:"acoustid_fingerprint"`
	AcoustIDID          *string  `json:"acoustid_id"`
	Album               *string  `json:"album"`
	Albumartist         *string  `json:"albumartist"`
	AlbumartistCredit   *string  `json:"albumartist_credit"`
	AlbumartistSort     *string  `json:"albumartist_sort"`
	Albumartists        []string `json:"albumartists"`
	AlbumartistsCredit  []string `json:"albumartists_credit"`
	AlbumartistsSort    []string `json:"albumartists_sort"`
	Albumdisambig       *string  `json:"albumdisambig"`
	Albumstatus         *string  `json:"albumstatus"`
	Albumtype           *string  `json:"albumtype"`
	Albumtypes          []string `json:"albumtypes"`
	Arrangers           []string `json:"arrangers"`
	Artist              *string  `json:"artist"`
	ArtistCredit        *string  `json:"artist_credit"`
	ArtistSort          *string  `json:"artist_sort"`
	Artists             []string `json:"artists"`
	ArtistsCredit       []string `json:"artists_credit"`
	ArtistsSort         []string `json:"artists_sort"`
	Asin                *string  `json:"asin"`
	Barcode             *string  `json:"barcode"`
	Bpm                 *int     `json:"bpm"`
	Catalognum          *string  `json:"catalognum"`
	Comments            *string  `json:"comments"`
	Comp                *bool    `json:"comp"`
	ComposerSort        *string  `json:"composer_sort"`
	Composers           []string `json:"composers"`
	Country             *string  `json:"country"`
	Year                *int     `json:"year"`
	Month               *int     `json:"month"`
	Day                 *int     `json:"day"`
	Disc                *int     `json:"disc"`
	Disctitle           *string  `json:"disctitle"`
	Disctotal           *int     `json:"disctotal"`
	Encoder             *string  `json:"encoder"`
	Genres              []string `json:"genres"`
	Grouping            *string  `json:"grouping"`
	InitialKey          *string  `json:"initial_key"`
	Isrc                *string  `json:"isrc"`
	Label               *string  `json:"label"`
	Language            *string  `json:"language"`
	Lyricists           []string `json:"lyricists"`
	Lyrics              *string  `json:"lyrics"`
	MBAlbumartistid     *string  `json:"mb_albumartistid"`
	MBAlbumartistids    []string `json:"mb_albumartistids"`
	MBAlbumid           *string  `json:"mb_albumid"`
	MBArtistid          *string  `json:"mb_artistid"`
	MBArtistids         []string `json:"mb_artistids"`
	MBReleasegroupid    *string  `json:"mb_releasegroupid"`
	MBReleasetrackid    *string  `json:"mb_releasetrackid"`
	MBTrackid           *string  `json:"mb_trackid"`
	MBWorkid            *string  `json:"mb_workid"`
	Media               *string  `json:"media"`
	OriginalYear        *int     `json:"original_year"`
	OriginalMonth       *int     `json:"original_month"`
	OriginalDay         *int     `json:"original_day"`
	R128AlbumGain       *float64 `json:"r128_album_gain"`
	R128TrackGain       *float64 `json:"r128_track_gain"`
	Remixers            []string `json:"remixers"`
	RgAlbumGain         *float64 `json:"rg_album_gain"`
	RgAlbumPeak         *float64 `json:"rg_album_peak"`
	RgTrackGain         *float64 `json:"rg_track_gain"`
	RgTrackPeak         *float64 `json:"rg_track_peak"`
	Script              *string  `json:"script"`
	Subtitle            *string  `json:"subtitle"`
	Title               *string  `json:"title"`
	Track               *int     `json:"track"`
	Tracktotal          *int     `json:"tracktotal"`
}

// EmptyString returns a present-but-empty value for a TagSet string
// member, a non-nil pointer distinct from the nil pointer that means
// absent (C2).
func EmptyString() *string {
	s := ""
	return &s
}

// EmptyList returns a present-but-empty value for a TagSet list
// member, a non-nil zero-length slice distinct from the nil slice
// that means absent (C2). go-conventions prefers a nil slice by
// default; this is the one case its own exception names, because the
// nil/empty distinction is the fact this function exists to carry.
func EmptyList() []string {
	return []string{}
}

// WriteOrder returns the 69 field identifiers in the field table's
// fixed write order.
func WriteOrder() []Field {
	order := make([]Field, len(fields))
	for i, f := range fields {
		order[i] = f.name
	}
	return order
}
